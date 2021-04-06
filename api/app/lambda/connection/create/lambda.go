// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package connectioncreatelambda

import (
	"fmt"

	"github.com/palchukovsky/ss"
	apiapp "github.com/palchukovsky/ss/api/app"
	"github.com/palchukovsky/ss/db"
	"github.com/palchukovsky/ss/ddb"
	ws "github.com/palchukovsky/ss/lambda/gateway/ws"
)

func Init(initService func(projectPackage string, params ss.ServiceParams)) {
	apiapp.Init(
		func() ws.Lambda { return lambda{db: ddb.GetClientInstance()} },
		func(projectPackage string) {
			initService(projectPackage, ss.ServiceParams{IsAWS: true})
		})
}

func Run() { apiapp.Run() }

////////////////////////////////////////////////////////////////////////////////

type lambda struct{ db ddb.Client }

func (lambda lambda) Execute(request ws.Request) error {
	err := lambda.db.Create(
		db.NewConnection(
			request.GetConnectionID(),
			request.GetUserID()))
	if err != nil {
		return fmt.Errorf(`failed to add connection: "%w"`, err)
	}
	request.Log().Debug("Connected.")
	return nil
}
