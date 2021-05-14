// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package main

import (
	"context"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/palchukovsky/ss"
)

type service struct {
	key    string
	secret string
	config ss.ServiceConfig
	log    ss.Log
	build  ss.Build
}

func newService(
	projectPackage string,
	key string,
	secret string,
	config ss.Config,
) ss.Service {
	return service{
		key:    key,
		secret: secret,
		config: config.SS.Service,
		log:    ss.NewLog(projectPackage, "test/gateway/install/create", config),
		build: ss.Build{
			Version:    "test",
			Commit:     "local",
			ID:         "local",
			Builder:    "local",
			Maintainer: "local",
		},
	}
}

func (s service) Log() ss.Log                         { return s.log }
func (service) Product() string                       { return "ss" }
func (service) Name() string                          { return "test" }
func (s service) Config() ss.ServiceConfig            { return s.config }
func (s service) Build() ss.Build                     { return s.build }
func (service) NewBuildEntityName(name string) string { return "test_" + name }

func (s service) StartLambda() {
	s.Log().Panic(ss.NewLogMsg("StartLambda is not implemented"))
}

func (s service) CompleteLambda(panicValue interface{}) {
	s.Log().Panic(ss.NewLogMsg("StartLambda is not implemented"))
}

func (s service) GetLambdaTimeout() <-chan time.Time {
	s.Log().Panic(ss.NewLogMsg("StartLambda is not implemented"))
	return nil
}

type credentials struct {
	key    string
	secret string
}

func (credentials credentials) Retrieve(
	context.Context,
) (aws.Credentials, error) {
	return aws.Credentials{
			AccessKeyID:     credentials.key,
			SecretAccessKey: credentials.secret,
		},
		nil
}

func (s service) NewAWSConfig() aws.Config {
	result, err := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion(s.Config().AWS.Region),
		config.WithCredentialsProvider(
			credentials{
				key:    s.key,
				secret: s.secret,
			}))
	if err != nil {
		s.Log().Panic(ss.NewLogMsg(`failed to load AWS config`).AddErr(err))
	}
	return result
}

func (s service) NewAWSSessionV1() *session.Session {
	s.Log().Panic(ss.NewLogMsg("AWS Config V1 is not implemented"))
	return nil
}
