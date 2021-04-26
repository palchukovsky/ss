// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package apiauth

import (
	"github.com/palchukovsky/ss"
	"github.com/palchukovsky/ss/lambda"
	rest "github.com/palchukovsky/ss/lambda/gateway/rest"
)

// Init initiates the auth-lambda.
func Init(
	newLambda func() rest.Lambda,
	initService func(projectPackage string),
) {
	initService("auth")
	defer ss.S.Log().CheckExit(recover())
	service = rest.NewService(newLambda())
}

// Run runs the API auth-lambda.
func Run() {
	defer ss.S.Log().CheckExit(recover())
	service.Start()
}

var service lambda.Service
