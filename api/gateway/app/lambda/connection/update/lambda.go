// Copyright 2021-2022, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package connectionupdatelambda

import (
	"github.com/palchukovsky/ss"
	"github.com/palchukovsky/ss/db"
	"github.com/palchukovsky/ss/ddb"
	ws "github.com/palchukovsky/ss/lambda/gateway/ws"
)

type lambda struct{ db ddb.Client }

func (lambda lambda) Execute(lambdaRequest ws.Request) error {
	var request request
	lambdaRequest.ReadRequest(&request)
	lambda.
		db.
		CreateOrReplace(
			db.NewDevice(
				request.Device,
				request.FCMToken,
				lambdaRequest.GetUserID(),
				lambdaRequest.ReadClientKey())).
		Request()
	lambdaRequest.Log().Debug(
		ss.NewLogMsg("updated").
			Add(request.Device).
			Add(request.FCMToken))
	lambdaRequest.Respond(newResponse())
	return nil
}
