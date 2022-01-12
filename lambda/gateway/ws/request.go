// Copyright 2021-2022, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package wsgatewaylambda

import (
	"encoding/json"

	"github.com/aws/aws-lambda-go/events"
	"github.com/palchukovsky/ss"
	"github.com/palchukovsky/ss/crypto"
	"github.com/palchukovsky/ss/db"
	"github.com/palchukovsky/ss/lambda"
	gate "github.com/palchukovsky/ss/lambda/gateway"
)

////////////////////////////////////////////////////////////////////////////////

// Request describes request to lambda.
type Request interface {
	ss.NoCopy

	Log() ss.LogStream

	GetConnectionID() ss.ConnectionID
	GetUserID() ss.UserID
	ReadClientInfo() gate.ClientInfo
	ReadClientKey() db.DeviceCryptoKey

	ReadRequest(interface{})

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
	log ss.LogSession,
) *request {

	user, err := ss.ParseUserID(
		getWSRequestHeaders(awsRequest)["principalId"].(string))
	if err != nil {
		log.Panic(ss.NewLogMsg("failed to parse user ID").AddErr(err))
	}

	logPrefix := ss.
		NewLogPrefix(func() []ss.LogMsgAttr { return nil }).
		Add(user).
		Add(ss.ConnectionID(awsRequest.RequestContext.ConnectionID)).
		AddRequestID(awsRequest.RequestContext.RequestID)

	return &request{
		Request: gate.NewRequest(
			gateway,
			log.NewSession(logPrefix),
			nil),
		AWSRequest: awsRequest,
		user:       user,
	}
}

func (request *request) GetConnectionID() ss.ConnectionID {
	return ss.ConnectionID(request.AWSRequest.RequestContext.ConnectionID)
}

func (request *request) GetUserID() ss.UserID { return request.user }

func (request *request) getHeader(name string) (string, bool) {
	result, has := getWSRequestHeaders(request.AWSRequest)[name]
	if !has {
		return "", has
	}
	return result.(string), true
}
func getWSRequestHeaders(
	source events.APIGatewayWebsocketProxyRequest,
) map[string]interface{} {
	return source.RequestContext.Authorizer.(map[string]interface{})
}

func (request *request) ReadClientInfo() gate.ClientInfo {
	result, err := gate.NewClientInfo(request.getHeader)
	if err != nil {
		request.Log().Panic(
			ss.NewLogMsg("failed to read client info").AddErr(err))
	}
	return result
}
func (request *request) ReadClientKey() crypto.AES128Key {
	result, err := gate.NewClientKey(request.getHeader)
	if err != nil {
		request.Log().Panic(
			ss.NewLogMsg("failed to read client key").AddErr(err))
	}
	return result

}

func (request *request) Respond(response interface{}) {
	request.Request.Respond(response)

	if ss.S.Config().IsExtraLogEnabled() {
		request.Log().Debug(ss.NewLogMsg("lambda response").AddResponse(response))
	}
}

func (request *request) ReadRequest(result interface{}) {
	if err := json.Unmarshal(request.readRequest()["d"], result); err != nil {
		request.Log().Panic(
			ss.NewLogMsg(`failed to parse request`).AddErr(err).AddDump(result))
	}
}

func (request *request) ReadRemoteID() *string {
	node, has := request.readRequest()["i"]
	if !has {
		return nil
	}
	var result string
	if err := json.Unmarshal(node, &result); err != nil {
		request.Log().Panic(
			ss.
				NewLogMsg(`failed to parse request remote ID`).
				AddErr(err))
	}
	return &result
}

func (request *request) readRequest() map[string]json.RawMessage {
	if request.rawRequest != nil {
		return request.rawRequest
	}

	if request.AWSRequest.Body == "" {
		request.rawRequest = map[string]json.RawMessage{}
	} else {
		err := request.UnmarshalRequest(
			request.AWSRequest.Body,
			&request.rawRequest)
		if err != nil {
			request.Log().Panic(
				ss.NewLogMsg(`failed to parse raw request`).AddErr(err))
		}
	}

	if ss.S.Config().IsExtraLogEnabled() {
		request.Log().Debug(
			ss.
				NewLogMsg("lambda request").
				AddRequest(request.rawRequest).
				AddRequest(request.AWSRequest))
	}

	return request.rawRequest
}

////////////////////////////////////////////////////////////////////////////////
