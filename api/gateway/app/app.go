// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package apiapp

import (
	"github.com/palchukovsky/ss"
	"github.com/palchukovsky/ss/lambda"
	ws "github.com/palchukovsky/ss/lambda/gateway/ws"
)

// Init initiates the app-lambda.
func Init(
	newLambda func() ws.Lambda,
	initService func(projectPackage string),
) {
	initService("app")
	defer ss.S.Log().CheckExit()
	service = ws.NewService(newLambda())
}

// Run runs the API app-lambda.
func Run() {
	defer ss.S.Log().CheckExit()
	service.Start()
}

var service lambda.Service
