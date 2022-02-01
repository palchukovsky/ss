// Copyright 2021-2022, the PriMedical project owners. All rights reserved.
// Please see the OWNERS file and website https://primedical.co/ for details.

package userdeletelambda

import (
	"github.com/palchukovsky/ss"
	"github.com/palchukovsky/ss/db"
	"github.com/palchukovsky/ss/ddb"
	ws "github.com/palchukovsky/ss/lambda/gateway/ws"
)

type lambda struct{ db ddb.Client }

func (lambda lambda) Execute(lambdaRequest ws.Request) error {
	isFound, err := db.DeleteUser(lambdaRequest.GetUserID(), lambda.db)
	if err != nil {
		return err
	}
	if isFound {
		lambdaRequest.Log().Info(ss.NewLogMsg("deleted"))
		lambdaRequest.Respond(newResponse())
	} else {
		lambdaRequest.Log().Debug(ss.NewLogMsg("no user record found"))
	}
	return nil
}
