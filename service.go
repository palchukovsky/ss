// Copyright 2021, the SS project owners. All rights reserved.
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

	// Go runs goroutine with all required checks.
	// Has to be used instead of native "go" each time when
	// goroutine executes business logic or could panic.
	Go(func())

	StartLambda(getFailInfo func() []LogMsgAttr)
	CompleteLambda(panicValue interface{})
	GetLambdaTimeout() <-chan time.Time

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

	lambdaTimeoutMutex     sync.Mutex
	lambdaTimeoutObservers *struct {
		Chans       []chan<- time.Time
		TimeoutTime *time.Time
	}

	fireabase unsafe.Pointer
}

func (service *service) Now() Time             { return Now() }
func (service *service) Log() Log              { return service.log }
func (service *service) Config() ServiceConfig { return service.config }
func (service *service) Build() Build          { return service.build }
func (service *service) Name() string          { return service.name }
func (service *service) Product() string       { return service.product }

func (service *service) Go(f func()) {
	go func() {
		defer func() { service.log.CheckExit(recover()) }()
		f()
	}()
}

func (service *service) Firebase() *firebase.App {
	result := atomic.LoadPointer(&service.fireabase)
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
		&service.fireabase,
		nil,
		unsafe.Pointer(app))
	if isSwapped {
		return app
	}
	return (*firebase.App)(atomic.LoadPointer(&service.fireabase))
}

func (service *service) StartLambda(getFailInfo func() []LogMsgAttr) {
	timeout := time.After(LambdaMaxRunTime)
	observers := &struct {
		Chans       []chan<- time.Time
		TimeoutTime *time.Time
	}{
		Chans: []chan<- time.Time{},
	}

	{
		service.lambdaTimeoutMutex.Lock()
		if service.lambdaTimeoutObservers == nil {
			service.lambdaTimeoutObservers = observers
		} else {
			observers = nil
		}
		service.lambdaTimeoutMutex.Unlock()
	}

	if observers == nil {
		service.Log().Panic(
			NewLogMsg("previous lambda isn't completed").AddAttrs(getFailInfo()))
	}

	go func() {

		value := <-timeout

		service.lambdaTimeoutMutex.Lock()
		defer service.lambdaTimeoutMutex.Unlock()
		// observers.Chans could be already not the same as not service has, but
		// mutex has to be the same as it could be the same also.

		var sync sync.WaitGroup
		sync.Add(1)
		go func() {
			for _, observerChan := range observers.Chans {
				observerChan <- value
			}
			sync.Done()
		}()

		if observers == service.lambdaTimeoutObservers {
			// Lambda isn't finished yet.
			service.Log().Error(
				NewLogMsg("%s lambda timeout", service.Name()).AddAttrs(getFailInfo()))
		}

		sync.Wait()

		observers.TimeoutTime = &value

	}()

}

func (service *service) CompleteLambda(panicValue interface{}) {
	service.lambdaTimeoutMutex.Lock()

	if service.lambdaTimeoutObservers == nil {
		service.lambdaTimeoutMutex.Unlock()
		logMessage := NewLogMsg("lambda isn't started")
		if panicValue != nil {
			service.Log().Error(logMessage)
			service.log.CheckExit(panicValue)
			return
		}
		service.Log().Panic(logMessage)
		return
	}

	service.lambdaTimeoutObservers.Chans = nil
	service.lambdaTimeoutObservers = nil

	service.lambdaTimeoutMutex.Unlock()

	service.log.CheckExit(panicValue)
}

func (service *service) GetLambdaTimeout() <-chan time.Time {
	result := make(chan time.Time, 1)
	service.lambdaTimeoutMutex.Lock()
	// If no timer - lambda has never started, so there is no any run time limit.
	if service.lambdaTimeoutObservers != nil {
		if service.lambdaTimeoutObservers.TimeoutTime == nil {
			service.lambdaTimeoutObservers.Chans = append(
				service.lambdaTimeoutObservers.Chans,
				result)
		} else {
			timeoutTime := *service.lambdaTimeoutObservers.TimeoutTime
			go func() { result <- timeoutTime }()
		}
	}
	service.lambdaTimeoutMutex.Unlock()
	return result
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
