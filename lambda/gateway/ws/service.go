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

func (service service) Start() { awslambda.Start(service.handle) }

func (service service) handle(request awsResquest) (awsResponse, error) {
	defer func() { ss.S.Log().CheckExit(recover()) }()

	lambdaRequest, err := newRequest(request, service.Gateway)
	if err != nil {
		return awsResponse{}, err
	}
	defer func() {
		lambdaRequest.Log().CheckPanic(
			recover(),
			func() *ss.LogMsg {
				return ss.
					NewLogMsg("request handling panic").
					AddRequest(lambdaRequest.AWSRequest)
			})
	}()

	if ss.S.Config().IsExtraLogEnabled() {
		lambdaRequest.Log().Debug(
			ss.NewLogMsg("lambda request").AddRequest(lambdaRequest.AWSRequest))
	}

	var response interface{}
	if err := service.lambda.Execute(&lambdaRequest); err != nil {
		lambdaRequest.Log().Error(
			ss.
				NewLogMsg(`lambda execution error`).
				AddErr(err).
				AddRequest(lambdaRequest.AWSRequest))
		response = newErrorResponseBody(lambdaRequest)
	} else {
		response = newSuccessResponseBody(lambdaRequest)
	}

	if ss.S.Config().IsExtraLogEnabled() {
		lambdaRequest.Log().Debug(
			ss.NewLogMsg("lambda response").AddResponse(response))
	}

	result := awsResponse{StatusCode: http.StatusOK}
	if response != nil {
		dump, err := json.Marshal(response)
		if err != nil {
			ss.S.Log().Panic(
				ss.
					NewLogMsg(`failed to marshal response`).
					AddErr(err).
					AddRequest(lambdaRequest.AWSRequest).
					AddResponse(response))
		}
		result.Body = string(dump)
	}

	return result, nil
}

func newSuccessResponseBody(request request) interface{} {
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

func newErrorResponseBody(request request) interface{} {
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
