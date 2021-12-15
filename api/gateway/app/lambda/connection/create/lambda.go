// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package connectioncreatelambda

import (
	"github.com/palchukovsky/ss"
	apiapp "github.com/palchukovsky/ss/api/gateway/app"
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

	client := request.ReadClientInfo()

	trans := ddb.NewWriteTrans()

	trans.Create(
		db.NewConnection(
			request.GetConnectionID(),
			request.GetUserID(),
			client.Version))

	if client.FCMToken != nil {
		trans.CreateOrReplace(
			db.NewDevice(
				client.Device,
				*client.FCMToken,
				request.GetUserID(),
				client.Key))
	}

	lambda.db.WriteWithResult(trans)

	request.Log().Debug(ss.NewLogMsg("connected").Add(client))
	return nil
}
