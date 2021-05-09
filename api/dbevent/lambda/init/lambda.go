// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package initlambda

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/palchukovsky/ss"
	apidbevent "github.com/palchukovsky/ss/api/dbevent"
	sslambda "github.com/palchukovsky/ss/lambda"
	dbeventlambda "github.com/palchukovsky/ss/lambda/dbevent"
)

type response struct {
	Build   string `json:"build"`
	Version string `json:"ver"`
}

// Lambda sends initial data for each new connection.
type lambda struct {
	gateway sslambda.Gateway
	message []byte
}

func (lambda lambda) Execute(request dbeventlambda.Request) error {
	for i, event := range request.GetEvents() {
		if err := lambda.execute(request, event); err != nil {
			return dbeventlambda.NewDBEventError(err, i, request)
		}
	}
	return nil
}

func (lambda lambda) execute(
	request dbeventlambda.Request,
	event events.DynamoDBEventRecord,
) error {
	if event.EventName != "INSERT" {
		return nil
	}

	connection := struct {
		ID ss.ConnectionID `json:"id"`
	}{}
	err := apidbevent.UnmarshalEventsDynamoDBAttributeValues(
		event.Change.Keys, &connection)
	if err != nil {
		return err
	}

	isSent, err := lambda.gateway.SendSerialized(connection.ID, lambda.message)
	if err != nil {
		return err
	}
	if !isSent {
		request.Log().Debug(ss.NewLogMsg("already disconnected").Add(connection.ID))
	}

	return nil
}
