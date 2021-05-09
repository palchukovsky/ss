// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package initlambda

import (
	"github.com/palchukovsky/ss"
	apidbevent "github.com/palchukovsky/ss/api/dbevent"
	sslambda "github.com/palchukovsky/ss/lambda"
	dbeventlambda "github.com/palchukovsky/ss/lambda/dbevent"
)

func Init(initService func(projectPackage string, params ss.ServiceParams)) {
	apidbevent.Init(
		func() dbeventlambda.Lambda {
			result := lambda{gateway: sslambda.NewGateway()}
			{
				build := ss.S.Build()
				var err error
				result.message, err = result.gateway.Serialize(
					response{Build: build.ID, Version: build.Version})
				if err != nil {
					ss.S.Log().Panic(ss.NewLogMsg(`failed to serialize`).AddErr(err))
				}
			}
			return result
		},
		func(projectPackage string) {
			initService(projectPackage, ss.ServiceParams{IsAWS: true})
		})
}

func Run() { apidbevent.Run() }
