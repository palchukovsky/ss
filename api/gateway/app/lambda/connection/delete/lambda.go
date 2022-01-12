// Copyright 2021-2022, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package connectiondeletelambda

import (
	"github.com/palchukovsky/ss"
	apiapp "github.com/palchukovsky/ss/api/gateway/app"
	"github.com/palchukovsky/ss/db"
	"github.com/palchukovsky/ss/ddb"
	ws "github.com/palchukovsky/ss/lambda/gateway/ws"
)

func Init(initService func(projectPackage string, params ss.ServiceParams)) {
	apiapp.Init(
		func() ws.Lambda {
			return lambda{db: ddb.GetClientInstance()}
		},
		func(projectPackage string) {
			initService(projectPackage, ss.ServiceParams{IsAWS: true})
		})
}

func Run() { apiapp.Run() }

type lambda struct{ db ddb.Client }

func (lambda lambda) Execute(request ws.Request) error {
	isSuccess := lambda.
		db.
		Delete(db.NewConnectionKey(request.GetConnectionID())).
		RequestWithResult().
		IsSuccess()
	if !isSuccess {
		request.Log().Warn(
			ss.
				NewLogMsg(`failed to find connection to delete`).
				Add(request.GetConnectionID()))
		return nil
	}
	request.Log().Debug(ss.NewLogMsg("disconnected"))
	return nil
}
