// Copyright 2021-2022, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package restgatewaylambda

import (
	"context"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/palchukovsky/ss"
	"github.com/palchukovsky/ss/lambda"
	gate "github.com/palchukovsky/ss/lambda/gateway"
)

////////////////////////////////////////////////////////////////////////////////

// Request describes request to lambda.
type Request interface {
	ss.NoCopy

	Log() ss.LogStream

	ReadClientVersion() string

	ReadRequest(interface{})

	Respond(interface{})
	RespondWithConflictError(interface{})
	RespondWithUnprocessableEntityError(interface{})
	RespondWithBadRequestError()
	RespondWithNotFoundError()
	RespondWithNotAcceptable()

	GetContext() context.Context
}

////////////////////////////////////////////////////////////////////////////////

type request struct {
	gate.Request
	AWSRequest events.APIGatewayProxyRequest
	StatusCode int
	context    context.Context
}

func newRequest(
	awsRequest events.APIGatewayProxyRequest,
	gateway lambda.Gateway,
	log ss.LogSession,
	context context.Context,
) request {
	return request{
		Request: gate.NewRequest(
			gateway,
			log.NewSession(
				ss.
					NewLogPrefix(func() []ss.LogMsgAttr { return nil }).
					AddRequestID(awsRequest.RequestContext.RequestID)),
			struct{}{}),
		AWSRequest: awsRequest,
		StatusCode: http.StatusOK,
		context:    context,
	}
}

func (request *request) GetContext() context.Context { return request.context }

func (request *request) ReadClientVersion() string {
	result, err := gate.NewClientVersion(
		func(name string) (string, bool) {
			result, has := request.AWSRequest.Headers[name]
			return result, has
		})
	if err != nil {
		request.Log().Panic(
			ss.NewLogMsg("failed to read client version").AddErr(err))
	}
	return result
}

func (request *request) ReadRequest(result interface{}) {
	err := request.UnmarshalRequest(request.AWSRequest.Body, result)
	if err != nil {
		request.Log().Panic(
			ss.NewLogMsg(`failed to parse request`).AddErr(err).AddDump(result))
	}
	if ss.S.Config().IsExtraLogEnabled() {
		request.Log().Debug(
			ss.
				NewLogMsg("lambda request").
				AddRequest(result).
				AddRequest(request.AWSRequest))
	}
}

func (request *request) Respond(response interface{}) {
	request.respond(http.StatusOK, response)
}

func (request *request) RespondWithConflictError(response interface{}) {
	request.respond(http.StatusConflict, response)
}

func (request *request) RespondWithUnprocessableEntityError(
	response interface{},
) {
	request.respond(http.StatusUnprocessableEntity, response)
}

func (request *request) RespondWithBadRequestError() {
	request.respond(http.StatusBadRequest, struct{}{})
}

func (request *request) RespondWithNotFoundError() {
	request.respond(http.StatusNotFound, struct{}{})
}

func (request *request) RespondWithNotAcceptable() {
	request.respond(http.StatusNotAcceptable, struct{}{})
}

func (request *request) respond(statusCode int, response interface{}) {
	request.Request.Respond(response)
	request.StatusCode = statusCode

	if ss.S.Config().IsExtraLogEnabled() {
		request.Log().Debug(
			ss.NewLogMsg("lambda response").
				AddVal("statusCode", statusCode).
				AddResponse(response))
	}
}

////////////////////////////////////////////////////////////////////////////////
