// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package connectiondeletelambda

import (
	"fmt"

	apiapp "github.com/palchukovsky/ss/api/app"
	"github.com/palchukovsky/ss/db"
	"github.com/palchukovsky/ss/ddb"
	ws "github.com/palchukovsky/ss/lambda/gateway/ws"
)

func Init(serviceInit func(projectPackage string)) {
	apiapp.Init(
		func() ws.Lambda {
			return lambda{db: ddb.GetClientInstance()}
		},
		serviceInit)
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
		request.Log().Warn(`Failed to find connection %q to delete.`,
			request.GetConnectionID())
		return nil
	}
	request.Log().Debug("Disconnected.")
	return nil
}
