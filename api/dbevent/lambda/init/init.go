// Copyright 2021-2022, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package initlambda

import (
	"github.com/palchukovsky/ss"
	apidbevent "github.com/palchukovsky/ss/api/dbevent"
)

func Init(initService func(projectPackage string, params ss.ServiceParams)) {
	apidbevent.Init(
		newLambda,
		func(projectPackage string) {
			initService(projectPackage, ss.ServiceParams{IsAWS: true})
		})
}

func Run() { apidbevent.Run() }
