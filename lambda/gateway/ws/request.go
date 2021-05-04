// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package wsgatewaylambda

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/palchukovsky/ss"
	"github.com/palchukovsky/ss/lambda"
	gate "github.com/palchukovsky/ss/lambda/gateway"
)

////////////////////////////////////////////////////////////////////////////////

// Request describes request to lambda.
type Request interface {
	Log() ss.LogStream

	GetConnectionID() ss.ConnectionID
	GetUserID() ss.UserID

	ReadRequest(interface{}) error

	Respond(interface{})
}

////////////////////////////////////////////////////////////////////////////////

type request struct {
	gate.Request
	AWSRequest events.APIGatewayWebsocketProxyRequest
	user       ss.UserID
	rawRequest map[string]json.RawMessage
}

func newRequest(
	awsRequest events.APIGatewayWebsocketProxyRequest,
	gateway lambda.Gateway,
) (request, error) {
	user, err := ss.ParseUserID(awsRequest.
		RequestContext.
		Authorizer.(map[string]interface{})["principalId"].(string))
	if err != nil {
		return request{}, err
	}
	logPrefix := ss.
		NewLogPrefix().
		AddUser(user).
		AddConnectionID(ss.ConnectionID(awsRequest.RequestContext.ConnectionID)).
		AddRequestID(awsRequest.RequestContext.RequestID)
	return request{
		Request:    gate.NewRequest(gateway, logPrefix, nil),
		AWSRequest: awsRequest,
		user:       user,
	}, nil
}

func (request request) GetConnectionID() ss.ConnectionID {
	return ss.ConnectionID(request.AWSRequest.RequestContext.ConnectionID)
}

func (request request) GetUserID() ss.UserID { return request.user }

func (request *request) ReadRequest(result interface{}) error {
	if err := json.Unmarshal(request.readRequest()["d"], result); err != nil {
		return fmt.Errorf(`failed to parse request argument: %w`, err)
	}
	return nil
}

func (request *request) ReadRemoteID() *string {
	node, has := request.readRequest()["i"]
	if !has {
		return nil
	}
	var result string
	if err := json.Unmarshal(node, &result); err != nil {
		ss.S.Log().Panic(
			ss.
				NewLogMsg(`failed to parse request remote ID`).
				AddErr(err).
				AddRequest(request.AWSRequest))
	}
	return &result
}

func (request *request) readRequest() map[string]json.RawMessage {
	if request.rawRequest != nil {
		return request.rawRequest
	}
	if request.AWSRequest.Body == "" {
		request.rawRequest = map[string]json.RawMessage{}
		return request.rawRequest
	}
	err := request.UnmarshalRequest(request.AWSRequest.Body, &request.rawRequest)
	if err != nil {
		ss.S.Log().Panic(
			ss.
				NewLogMsg(`failed to parse raw request`).
				AddErr(err).
				AddRequest(request.AWSRequest))
	}
	return request.rawRequest
}

////////////////////////////////////////////////////////////////////////////////
