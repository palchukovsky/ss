// Copyright 2021-2022, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package initlambda

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/palchukovsky/ss"
	apidbevent "github.com/palchukovsky/ss/api/dbevent"
	sslambda "github.com/palchukovsky/ss/lambda"
	dbeventlambda "github.com/palchukovsky/ss/lambda/dbevent"
)

type errorCode string

const (
	errorCodeVersionIsNotActual = "ver-not-actual"
)

type response struct {
	Build     string    `json:"build"`
	Version   string    `json:"ver"`
	ErrorCode errorCode `json:"error,omitempty"`
	Message   string    `json:"message,omitempty"`
}

func newDefaultResponse() response {
	build := ss.S.Build()
	return response{Build: build.ID, Version: build.Version}
}

////////////////////////////////////////////////////////////////////////////////

// Lambda sends initial data for each new connection.
type lambda struct {
	gateway        sslambda.Gateway
	defaultMessage []byte
}

func newLambda() dbeventlambda.Lambda {
	result := lambda{gateway: sslambda.NewGateway()}
	result.defaultMessage = result.gateway.Serialize(newDefaultResponse())
	return result
}

func (lambda lambda) Execute(request dbeventlambda.Request) error {
	for _, event := range request.GetEvents() {
		lambda.execute(request, event)
	}
	return nil
}

func (lambda lambda) execute(
	request dbeventlambda.Request,
	event events.DynamoDBEventRecord,
) {
	if event.EventName != "INSERT" {
		return
	}

	connection := struct {
		ID ss.ConnectionID `json:"id"`
		// Required by BUZZ-78, but disabled to don't send full record until
		// version control required (see substring BUZZ-78 for other details):
		// Version string `json:"ver"`
	}{}
	apidbevent.UnmarshalEventsDynamoDBAttributeValues(
		// -------------------------------------------------------------------------
		/*
			Required by BUZZ-78, but disabled to don't send full record until
			version control required (see substring BUZZ-78 for other details):

			event.Change.NewImage,
		*/
		event.Change.Keys,
		// -------------------------------------------------------------------------
		&connection)

	// ---------------------------------------------------------------------------
	/*
		Required by BUZZ-78, but disabled to don't send full record until
		version control required (see substring BUZZ-78 for other details):

		isActualVersion, err := apigateway.CheckClientVersionActuality(
			connection.Version)
		if err != nil {
			return err
		}
	*/
	isActualVersion := true
	// ---------------------------------------------------------------------------

	gateway := lambda.gateway.NewSessionGatewaySendSession(request.Log())
	if isActualVersion {
		gateway.SendSerialized(connection.ID, lambda.defaultMessage)
	} else {
		message := newDefaultResponse()
		message.ErrorCode = errorCodeVersionIsNotActual
		gateway.Send(connection.ID, message)
	}

	// Init isn't worry is dbevent executed or no, so it doesn't check stat,
	// just waits until the message will be sent:
	gateway.Close()
}

////////////////////////////////////////////////////////////////////////////////
