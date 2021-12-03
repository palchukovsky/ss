// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package wsgatewaylambda

import (
	"encoding/json"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	awslambda "github.com/aws/aws-lambda-go/lambda"
	"github.com/palchukovsky/ss"
	"github.com/palchukovsky/ss/lambda"
	gate "github.com/palchukovsky/ss/lambda/gateway"
)

// NewService creates new lambda service instance for lambda which works with WebSocket route.
func NewService(lambda Lambda) lambda.Service {
	return &service{
		Service: gate.NewService(),
		lambda:  lambda,
	}
}

type awsResquest = events.APIGatewayWebsocketProxyRequest
type awsResponse = events.APIGatewayProxyResponse

type service struct {
	gate.Service
	lambda Lambda
}

func (service service) Start() {
	awslambda.Start(
		func(request awsResquest) (awsResponse, error) {
			return service.handle(request), nil
		})
}

func (service service) handle(request awsResquest) awsResponse {
	ss.S.StartLambda(
		func() []ss.LogMsgAttr { return ss.NewLogMsgAttrRequestDumps(request) })
	defer func() { ss.S.CompleteLambda(recover()) }()

	log := ss.S.Log().NewSession(
		ss.NewLogPrefix(
			func() []ss.LogMsgAttr { return ss.NewLogMsgAttrRequestDumps(request) }))
	defer func() { log.CheckPanic(recover(), "request handling panic") }()

	lambdaRequest := newRequest(request, service.Gateway, log)

	var response interface{}
	if err := service.lambda.Execute(lambdaRequest); err != nil {
		lambdaRequest.Log().Error(
			ss.
				NewLogMsg(`lambda execution error`).
				AddErr(err).
				AddRequest(lambdaRequest.AWSRequest))
		response = newErrorResponseBody(lambdaRequest)
	} else {
		response = newSuccessResponseBody(lambdaRequest)
	}

	result := awsResponse{StatusCode: http.StatusOK}
	if response != nil {
		dump, err := json.Marshal(response)
		if err != nil {
			ss.S.Log().Panic(
				ss.
					NewLogMsg(`failed to marshal response`).
					AddErr(err).
					AddResponse(response))
		}
		result.Body = string(dump)
	}

	return result
}

func newSuccessResponseBody(request *request) interface{} {
	id := request.ReadRemoteID()
	if id == nil {
		return nil
	}
	return struct {
		ID   string      `json:"i"`
		Data interface{} `json:"d"`
	}{
		ID:   *id,
		Data: request.ResponseBody,
	}
}

func newErrorResponseBody(request *request) interface{} {
	id := request.ReadRemoteID()
	if id == nil {
		return nil
	}
	return struct {
		ID    string `json:"i"`
		Error string `json:"e"`
	}{
		ID:    *id,
		Error: "server error",
	}
}
