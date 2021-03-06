// Copyright 2021-2022, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package apidbevent

import (
	"github.com/palchukovsky/ss"
	"github.com/palchukovsky/ss/lambda"
	dbeventlambda "github.com/palchukovsky/ss/lambda/dbevent"
)

// Init initiates the dbevent-lambda.
func Init(
	newLambda func() dbeventlambda.Lambda,
	initService func(projectPackage string),
) {
	initService("dbevent")
	defer func() { ss.S.Log().CheckExit(recover()) }()
	service = dbeventlambda.NewService(newLambda())
}

// Run runs the API dbevent-lambda.
func Run() {
	defer func() { ss.S.Log().CheckExit(recover()) }()
	service.Start()
}

var service lambda.Service
