// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package main

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/palchukovsky/ss"
)

type service struct {
	config ss.ServiceConfig
	log    ss.ServiceLog
	build  ss.Build
}

func newService(projectPackage string, config ss.ServiceConfig) service {

	return service{
		config: config,
		log:    ss.NewServiceDevLog(projectPackage, "test/gateway/install/create"),
		build: ss.Build{
			Version:    "test",
			Commit:     "local",
			ID:         "local",
			Builder:    "local",
			Maintainer: "local",
		},
	}
}

func (s service) Log() ss.ServiceLog                  { return s.log }
func (service) Product() string                       { return "ss" }
func (service) Name() string                          { return "test" }
func (s service) Config() ss.ServiceConfig            { return s.config }
func (s service) Build() ss.Build                     { return s.build }
func (service) NewBuildEntityName(name string) string { return "test_" + name }

func (s service) NewAWSConfig() aws.Config {
	result, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		s.Log().Panic(`Failed to load AWS config: "%v".`, err)
	}
	return result
}

func (s service) NewAWSSessionV1() *session.Session {
	s.Log().Panic("AWS Config V1 is not implemented")
	return nil
}
