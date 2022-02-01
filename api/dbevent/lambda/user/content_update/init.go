// Copyright 2021-2022, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package usercontentupdatelambda

import (
	"github.com/palchukovsky/ss"
	apidbevent "github.com/palchukovsky/ss/api/dbevent"
	dbeventlambda "github.com/palchukovsky/ss/lambda/dbevent"
)

func Init(
	initService func(projectPackage string, params ss.ServiceParams),
	newNewUserUpdater *UpdaterFactory,
	newUpdatedUserUpdater *UpdaterFactory,
	newDeletedUserUpdater *UpdaterFactory,
) {
	apidbevent.Init(
		func() dbeventlambda.Lambda {
			return newLambda(
				newNewUserUpdater,
				newUpdatedUserUpdater,
				newDeletedUserUpdater)
		},
		func(projectPackage string) {
			initService(projectPackage, ss.ServiceParams{IsAWS: true})
		})
}

func Run() { apidbevent.Run() }
