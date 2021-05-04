// Copyright 2021, the SS project owners. All rights reserved.
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
	Log() ss.LogStream

	ReadRequest(interface{}) error

	Respond(interface{})
	RespondWithConflictError(interface{})
	RespondWithUnprocessableEntityError(interface{})
	RespondWithBadRequestError()
	RespondWithNotFoundError()

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
	context context.Context,
) request {
	return request{
		Request: gate.NewRequest(
			gateway,
			ss.NewLogPrefix().AddRequestID(awsRequest.RequestContext.RequestID),
			struct{}{}),
		AWSRequest: awsRequest,
		StatusCode: http.StatusOK,
		context:    context,
	}
}

func (request *request) GetContext() context.Context { return request.context }

func (request *request) ReadRequest(result interface{}) error {
	return request.UnmarshalRequest(request.AWSRequest.Body, result)
}

func (request *request) Respond(response interface{}) {
	request.Request.Respond(response)
	request.StatusCode = http.StatusOK
}

func (request *request) RespondWithConflictError(response interface{}) {
	request.Request.Respond(response)
	request.StatusCode = http.StatusConflict
}

func (request *request) RespondWithUnprocessableEntityError(
	response interface{}) {
	request.Request.Respond(response)
	request.StatusCode = http.StatusUnprocessableEntity
}

func (request *request) RespondWithBadRequestError() {
	request.Request.Respond(struct{}{})
	request.StatusCode = http.StatusBadRequest
}

func (request *request) RespondWithNotFoundError() {
	request.Request.Respond(struct{}{})
	request.StatusCode = http.StatusNotFound
}

////////////////////////////////////////////////////////////////////////////////
