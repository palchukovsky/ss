// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package connectiondeletelambda

import (
	"fmt"

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
	isFound, err := lambda.
		db.
		Delete(db.NewConnectionKey(request.GetConnectionID())).
		Request()
	if err != nil {
		return fmt.Errorf(`failed to delete connection %q: "%w"`,
			request.GetConnectionID(), err)
	}
	if !isFound {
		request.Log().Warn(
			ss.
				NewLogMsg(`failed to find connection to delete`).
				Add(request.GetConnectionID()))
		return nil
	}
	request.Log().Debug(ss.NewLogMsg("disconnected"))
	return nil
}
