// Copyright 2021-2022, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ss

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"
	"unsafe"

	firebase "firebase.google.com/go"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go/aws/session"
	"google.golang.org/api/option"
)

////////////////////////////////////////////////////////////////////////////////

// Service is the root interface of the service.
type Service interface {
	NoCopy

	Now() Time

	Log() Log
	Product() string
	Name() string
	Config() ServiceConfig
	Build() Build

	StartLambda(getFailInfo func() []LogMsgAttr)
	CompleteLambda(panicValue interface{})
	SubscribeForLambdaTimeout() <-chan struct{}

	NewBuildEntityName(name string) string

	NewAWSConfig() aws.Config
	NewAWSSessionV1() *session.Session

	Firebase() *firebase.App
}

////////////////////////////////////////////////////////////////////////////////

// ServiceParams is a set of service parameters.
type ServiceParams struct {
	IsAWS      bool
	IsFirebase bool
	IsAuth     bool
}

// NewService creates new service instance.
func NewService(
	product string,
	projectPackage string,
	params ServiceParams,
	configWrapper ConfigWrapper,
) Service {

	name := os.Args[0]
	name = name[strings.LastIndexAny(name, `/\\`)+1:]

	rand.Seed(Now().Get().UnixNano())

	config := configWrapper.GetSSPtr()
	{
		config.SS.Service.AWS.AccessKey.IsUsed = params.IsAWS
		config.SS.Service.PrivateKey.RSA.IsUsed = params.IsAuth
		config.SS.Service.Firebase.IsUsed = params.IsFirebase

		file, err := ioutil.ReadFile("config.json")
		if err != nil {
			log.Fatalf(`Failed to read config file: "%v".`, err)
		}
		if err := json.Unmarshal(file, configWrapper.GetBasePtr()); err != nil {
			log.Fatalf(`Failed to parse config file: "%v".`, err)
		}

		if !config.SS.Build.IsProd() {
			config.SS.Build.ID += "-" + config.SS.Build.Version
		}

		{
			varName := "SS_AWS_LAMBDA_TIMEOUT"
			varVal := os.Getenv(varName)
			if varVal == `` {
				config.SS.Service.AWS.LambdaTimeout = 3 * time.Second
			} else {
				seconds, err := strconv.Atoi(varVal)
				if err != nil {
					log := NewLogMsg(
						`failed to parse environment variable "%s" with lambda timeout "%s"`,
						varName,
						varVal).
						AddErr(err)
					S.Log().Panic(log)
				}
				config.SS.Service.AWS.LambdaTimeout =
					time.Duration(seconds) * time.Second
			}
		}

	}

	return &service{
		name:    name,
		product: product,
		config:  config.SS.Service,
		log:     NewLog(projectPackage, name, *config),
		build:   config.SS.Build,
	}
}

type service struct {
	NoCopyImpl

	name    string
	product string
	config  ServiceConfig
	log     Log
	build   Build

	lambdaTimeout lambdaTimeout

	firebase unsafe.Pointer
}

func (service *service) Now() Time             { return Now() }
func (service *service) Log() Log              { return service.log }
func (service *service) Config() ServiceConfig { return service.config }
func (service *service) Build() Build          { return service.build }
func (service *service) Name() string          { return service.name }
func (service *service) Product() string       { return service.product }

func (service *service) Firebase() *firebase.App {
	result := atomic.LoadPointer(&service.firebase)
	if result != nil {
		return (*firebase.App)(result)
	}
	app, err := firebase.NewApp(
		context.Background(),
		nil,
		option.WithCredentialsJSON(service.Config().Firebase.GetJSON()))
	if err != nil {
		service.Log().Panic(NewLogMsg(`failed to init Firebase`).AddErr(err))
	}
	isSwapped := atomic.CompareAndSwapPointer(
		&service.firebase,
		nil,
		unsafe.Pointer(app))
	if isSwapped {
		return app
	}
	return (*firebase.App)(atomic.LoadPointer(&service.firebase))
}

func (service *service) StartLambda(getFailInfo func() []LogMsgAttr) {
	timeout := service.Config().AWS.LambdaTimeout
	timeout -= (time.Duration(500) * time.Millisecond)

	service.lambdaTimeout.Start(timeout, getFailInfo)
}

func (service *service) CompleteLambda(panicValue interface{}) {

	// @palchukovsky: It has to be called before the lambda timeout watcher stop
	// because without it impossible to catch a log-sync timeout
	// (and many other timeouts).
	service.log.CheckExit(panicValue)

	service.lambdaTimeout.Cancel()
}

func (service *service) SubscribeForLambdaTimeout() <-chan struct{} {
	return service.lambdaTimeout.Subscribe()
}

func (service *service) NewBuildEntityName(name string) string {
	return fmt.Sprintf("%s_%s_%s",
		service.Product(),
		service.Build().Version,
		name)
}

func (service *service) NewAWSConfig() aws.Config {
	result, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		service.Log().Panic(NewLogMsg(`failed to load AWS config`).AddErr(err))
	}
	return result
}

func (service *service) NewAWSSessionV1() *session.Session {
	result, err := session.NewSessionWithOptions(session.Options{
		SharedConfigState: session.SharedConfigEnable,
	})
	if err != nil {
		service.Log().Panic(NewLogMsg(`failed to load AWS v1 config`).AddErr(err))
	}
	return result
}

////////////////////////////////////////////////////////////////////////////////

type lambdaTimeout struct {
	NoCopyImpl
	mutex      sync.RWMutex
	cancelChan chan struct{}
	waiter     *lambdaTimeoutWaiter
}

func (t *lambdaTimeout) Start(
	timeout time.Duration,
	getFailInfo func() []LogMsgAttr,
) {
	// No sync, start-cancel are always synchronized.
	// But subscribing is not (see below).

	if t.cancelChan != nil {
		S.Log().Panic(
			NewLogMsg("previous lambda isn't completed").AddAttrs(getFailInfo()))
	}

	t.cancelChan = make(chan struct{}, 1)

	waiter := newLambdaTimeoutWaiter(t.cancelChan, timeout, getFailInfo)

	t.mutex.Lock()
	defer t.mutex.Unlock()

	t.waiter = waiter
}

func (t *lambdaTimeout) Cancel() {
	// No sync, start-cancel are always synchronized.

	if t.cancelChan == nil {
		S.Log().Panic(NewLogMsg(`lambda timeout is not started`))
	}

	t.cancelChan <- struct{}{}
	t.cancelChan = nil

	t.waiter.Expire()
}

func (t *lambdaTimeout) Subscribe() <-chan struct{} {
	result := make(chan struct{}, 1)

	isObservationActive := false
	defer func() {
		if isObservationActive {
			return
		}
		result <- struct{}{}
	}()

	t.mutex.RLock()
	waiter := t.waiter
	t.mutex.RUnlock()

	if waiter == nil {
		// Means some lambda doesn't have "StartLambda" call,
		// so it doesn't have timeout.
		isObservationActive = true
		return result
	}

	isObservationActive = waiter.Subscribe(result)
	return result
}

type lambdaTimeoutWaiter struct {
	NoCopyImpl
	mutex     sync.Mutex
	observers []chan<- struct{}
	isExpired bool
}

func newLambdaTimeoutWaiter(
	cancelChan chan struct{},
	timeout time.Duration,
	getFailInfo func() []LogMsgAttr,
) *lambdaTimeoutWaiter {
	result := &lambdaTimeoutWaiter{}
	go result.wait(cancelChan, timeout, getFailInfo)
	return result
}

func (w *lambdaTimeoutWaiter) wait(
	cancelChan chan struct{},
	timeout time.Duration,
	getFailInfo func() []LogMsgAttr,
) {
	defer func() {
		w.mutex.Lock()
		defer w.mutex.Unlock()

		w.isExpired = true

		for _, observer := range w.observers {
			observer <- struct{}{}
		}
	}()

	select {
	case <-cancelChan:
		// Lambda is completed before its timeout.
		break
	case timeoutTime := <-time.After(timeout):
		// Lambda has reached its timeout.
		{
			message := NewLogMsg(
				"%s lambda timeout %f seconds on %s",
				S.Name(),
				timeout.Seconds(),
				timeoutTime)
			S.Log().Error(message.AddAttrs(getFailInfo()))
		}
		break
	}
}

func (w *lambdaTimeoutWaiter) Subscribe(observerChan chan<- struct{}) bool {
	w.mutex.Lock()
	defer w.mutex.Unlock()

	if w.isExpired {
		return false
	}

	w.observers = append(w.observers, observerChan)
	return true
}

func (w *lambdaTimeoutWaiter) Expire() {
	w.mutex.Lock()
	w.isExpired = true
	w.mutex.Unlock()
}

////////////////////////////////////////////////////////////////////////////////
