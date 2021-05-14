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
	lambda "github.com/palchukovsky/ss/api/gateway/install/lambda/create"
)

////////////////////////////////////////////////////////////////////////////////

func init() {

	config := ss.Config{}

	lambda.Init(
		func(projectPackage string, params ss.ServiceParams) {
			ss.Set(
				newService(
					projectPackage,
					"",
					"",
					config))
		})
}

func main() {
	defer func() { ss.S.Log().CheckExit(nil) }()
	ss.S.Log().Debug(ss.NewLogMsg("first debug"))
	ss.S.Log().Info(ss.NewLogMsg("second debug"))
}

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
	config ss.Config) service {

	return service{
		key:    key,
		secret: secret,
		config: config.SS.Service,
		log:    ss.NewLog(projectPackage, "asdAsd", config),
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

func (service) StartLambda()                       {}
func (service) CompleteLambda(interface{})         {}
func (service) GetLambdaTimeout() <-chan time.Time { return nil }

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
	result, _ := config.LoadDefaultConfig(
		context.TODO(),
		config.WithRegion(s.Config().AWS.Region),
		config.WithCredentialsProvider(
			credentials{
				key:    s.key,
				secret: s.secret,
			}))
	// if err != nil {
	//s.Log().Panic(`Failed to load AWS config: "%v".`, err)
	// }
	return result
}

func (s service) NewAWSSessionV1() *session.Session {
	// s.Log().Panic("AWS Config V1 is not implemented")
	return nil
}
