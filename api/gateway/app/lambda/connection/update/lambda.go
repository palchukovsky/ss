// Copyright 2021-2022, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package connectionupdatelambda

import (
	"github.com/palchukovsky/ss"
	apiapp "github.com/palchukovsky/ss/api/gateway/app"
	"github.com/palchukovsky/ss/db"
	"github.com/palchukovsky/ss/ddb"
	ws "github.com/palchukovsky/ss/lambda/gateway/ws"
)

type request struct {
	Device   ss.DeviceID                    `json:"device"`
	FCMToken ss.FirebaseCloudMessagingToken `json:"fcm"`
}

type response struct{}

func newResponse() response { return response{} }

////////////////////////////////////////////////////////////////////////////////

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
