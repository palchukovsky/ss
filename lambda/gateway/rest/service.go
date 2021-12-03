// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package restgatewaylambda

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	awslambda "github.com/aws/aws-lambda-go/lambda"
	"github.com/palchukovsky/ss"
	"github.com/palchukovsky/ss/lambda"
	gate "github.com/palchukovsky/ss/lambda/gateway"
)

// NewService creates new lambda service instance for lambda which works with REST route.
func NewService(lambda Lambda) lambda.Service {
	return &service{
		Service: gate.NewService(),
		lambda:  lambda,
		context: context.Background(),
	}
}

type awsRequest = events.APIGatewayProxyRequest
type awsResponse = events.APIGatewayProxyResponse

type service struct {
	gate.Service
	lambda  Lambda
	context context.Context
}

func (service service) Start() {
	awslambda.StartWithContext(
		service.context,
		func(request awsRequest) (awsResponse, error) {
			return service.handle(request), nil
		})
}

func (service service) handle(request awsRequest) awsResponse {
	ss.S.StartLambda(
		func() []ss.LogMsgAttr { return ss.NewLogMsgAttrRequestDumps(request) })
	defer func() { ss.S.CompleteLambda(recover()) }()

	log := ss.S.Log().NewSession(
		ss.NewLogPrefix(
			func() []ss.LogMsgAttr { return ss.NewLogMsgAttrRequestDumps(request) }))

	lambdaRequest := newRequest(request, service.Gateway, log, service.context)
	defer func() {
		lambdaRequest.Log().CheckPanic(recover(), "panic at request handling")
	}()

	if err := service.validateRequest(request); err != nil {
		lambdaRequest.Log().Warn(
			ss.NewLogMsg(`request validation failed`).AddErr(err))
		return awsResponse{StatusCode: http.StatusBadRequest}
	}

	if err := service.lambda.Execute(&lambdaRequest); err != nil {
		lambdaRequest.Log().Panic(
			ss.NewLogMsg(`lambda execution error`).AddErr(err))
	}

	result := awsResponse{
		StatusCode: lambdaRequest.StatusCode,
		// Headers have to be in lowercase for better compression.
		// Also, Cloudflare converts it in lower case.
		Headers: map[string]string{
			"content-type": "application/json; charset=utf-8",
		},
	}

	if lambdaRequest.ResponseBody != nil {
		dump, err := json.Marshal(lambdaRequest.ResponseBody)
		if err != nil {
			lambdaRequest.Log().Panic(
				ss.
					NewLogMsg(`failed to marshal response`).
					AddErr(err).
					AddResponse(lambdaRequest.ResponseBody))
		}
		result.Body = string(dump)
	}

	return result
}

func (service service) validateRequest(request awsRequest) error {
	// The temporary solution to be sure that AWS already validated the request,
	// see for details: https://buzzplace.atlassian.net/browse/Buzz-85
	// Headers have to be in lowercase for better compression. Also, Cloudflare
	// converts it in lower case.
	contentType, has := request.Headers["content-type"]
	if !has {
		return fmt.Errorf("content type is not set")
	}
	if contentType != "application/json; charset=utf-8" {
		return fmt.Errorf("invalid content type %q", contentType)
	}
	return nil
}
