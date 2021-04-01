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
	Log() ss.ServiceLog

	GetConnectionID() lambda.ConnectionID
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
	logPrefix := fmt.Sprintf("%s.%s.%s",
		user,
		awsRequest.RequestContext.ConnectionID,
		awsRequest.RequestContext.RequestID)
	return request{
		Request:    gate.NewRequest(gateway, logPrefix, nil),
		AWSRequest: awsRequest,
		user:       user,
	}, nil
}

func (request request) GetConnectionID() lambda.ConnectionID {
	return lambda.ConnectionID(request.AWSRequest.RequestContext.ConnectionID)
}

func (request request) GetUserID() ss.UserID { return request.user }

func (request *request) ReadRequest(result interface{}) error {
	if err := json.Unmarshal(request.readRequest()["d"], result); err != nil {
		return fmt.Errorf(
			`failed to parse request argument: "%w". Type %q. Dump: %q`,
			err, ss.GetTypeName(result), request.AWSRequest.Body)
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
		ss.S.Log().Panic(`Failed to parse request remote ID: "%v. Dump: %s`,
			err, ss.Dump(request.AWSRequest))
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
		ss.S.Log().Panic(`Failed to parse raw request: "%v". Request dump: %s`,
			err, ss.Dump(request.AWSRequest))
	}
	return request.rawRequest
}

////////////////////////////////////////////////////////////////////////////////
