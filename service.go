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
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go/aws/session"
)

////////////////////////////////////////////////////////////////////////////////

// Service is the root interface of the service.
type Service interface {
	Log() Log
	Product() string
	Name() string
	Config() ServiceConfig
	Build() Build

	StartLambda()
	GetLambdaTimeout() <-chan time.Time

	NewBuildEntityName(name string) string

	NewAWSConfig() aws.Config
	NewAWSSessionV1() *session.Session
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

	rand.Seed(Now().UnixNano())

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
}

func (service *service) Log() Log              { return service.log }
func (service *service) Config() ServiceConfig { return service.config }
func (service *service) Build() Build          { return service.build }
func (service *service) Name() string          { return service.name }
func (service *service) Product() string       { return service.product }

func (service *service) StartLambda() {
	timeout := time.After(LambdaMaxRunTime)
	observers := &struct {
		Chans       []chan<- time.Time
		TimeoutTime *time.Time
	}{
		Chans: []chan<- time.Time{},
	}

	service.lambdaTimeoutMutex.Lock()
	service.lambdaTimeoutObservers = observers
	service.lambdaTimeoutMutex.Unlock()

	go func() {
		value := <-timeout
		service.lambdaTimeoutMutex.Lock()
		defer service.lambdaTimeoutMutex.Unlock()
		// observers.Chans could be already not the same as not service has, but
		// mutex has to be the same as it could be the same also.
		for _, observerChan := range observers.Chans {
			observerChan <- value
		}
		observers.TimeoutTime = &value
	}()

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
