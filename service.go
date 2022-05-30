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

	lambdaTimeout *lambdaTimeout

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
	if service.lambdaTimeout != nil {
		service.Log().Panic(
			NewLogMsg("previous lambda isn't completed").AddAttrs(getFailInfo()))
	}

	timeout := service.Config().AWS.LambdaTimeout
	timeout -= (time.Duration(500) * time.Millisecond)
	service.lambdaTimeout = newLambdaTimeout(timeout, getFailInfo)
}

func (service *service) CompleteLambda(panicValue interface{}) {

	// @palchukovsky: It has to be called before the lambda timeout watcher stop
	// because without it impossible to catch a log-sync timeout
	// (and many other timeouts).
	service.log.CheckExit(panicValue)

	if service.lambdaTimeout == nil {
		logMessage := NewLogMsg("lambda isn't started, failed to complete")
		if panicValue != nil {
			service.Log().Error(logMessage)
			service.log.CheckExit(panicValue)
			return
		}
		service.Log().Panic(logMessage)
		return
	}

	service.lambdaTimeout.Cancel()
	service.lambdaTimeout = nil
}

func (service *service) SubscribeForLambdaTimeout() <-chan struct{} {
    timeout := service.lambdaTimeout
	if timeout == nil {
		service.Log().Panic(NewLogMsg("lambda isn't started, failed to subscribe"))
	}
	return timeout.Subscribe()
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

	mutex sync.Mutex

	timeout     time.Duration
	getFailInfo func() []LogMsgAttr

	timeoutChan <-chan time.Time
	cancelChan  chan struct{}

	observers []chan<- struct{}
	isExpired bool
}

func newLambdaTimeout(
	timeout time.Duration,
	getFailInfo func() []LogMsgAttr,
) *lambdaTimeout {
	result := &lambdaTimeout{
		timeout:     timeout,
		timeoutChan: time.After(timeout),
		cancelChan:  make(chan struct{}, 1),
		getFailInfo: getFailInfo,
	}

	go result.wait()

	return result
}

func (timeout *lambdaTimeout) Cancel() { timeout.cancelChan <- struct{}{} }

func (timeout *lambdaTimeout) Subscribe() <-chan struct{} {
	result := make(chan struct{}, 1)

	timeout.mutex.Lock()
	if !timeout.isExpired {
		timeout.observers = append(timeout.observers, result)
	} else {
		result <- struct{}{}
	}
	timeout.mutex.Unlock()

	return result
}

func (timeout *lambdaTimeout) wait() {
	var timeoutTime time.Time
	select {
	case <-timeout.cancelChan:
		// Lambda is completed before its timeout.
		return
	case timeoutTime = <-timeout.timeoutChan:
		// Lambda has reached its timeout.
		break
	}

	{
		message := NewLogMsg(
			"%s lambda timeout %f seconds on %s",
			S.Name(),
			timeout.timeout.Seconds(),
			timeoutTime)
		S.Log().Error(message.AddAttrs(timeout.getFailInfo()))
	}

	timeout.mutex.Lock()

	timeout.isExpired = true

	for _, observer := range timeout.observers {
		observer <- struct{}{}
	}

	timeout.mutex.Unlock()
}

////////////////////////////////////////////////////////////////////////////////
