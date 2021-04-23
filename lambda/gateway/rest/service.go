// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package restgatewaylambda

import (
	"context"
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
	awslambda.StartWithContext(service.context, service.handle)
}

func (service service) handle(request awsRequest) (awsResponse, error) {

	lambdaRequest := newRequest(request, service.Gateway, service.context)
	defer lambdaRequest.Log().CheckExit()
	if ss.S.Config().IsExtraLogEnabled() {
		lambdaRequest.Log().Debug(
			"Lambda request dump: %s",
			ss.Dump(lambdaRequest.AWSRequest))
	}
	if err := service.validateRequest(request); err != nil {
		lambdaRequest.Log().Warn(
			`Request validation failed: "%v". Request dump: %s`, err,
			ss.Dump(lambdaRequest.AWSRequest))
		return awsResponse{StatusCode: http.StatusBadRequest}, nil
	}

	if err := service.lambda.Execute(&lambdaRequest); err != nil {
		lambdaRequest.Log().Error(
			`Lambda execution error: "%v". Request dump: %s`,
			err,
			ss.Dump(lambdaRequest.AWSRequest))
		return awsResponse{StatusCode: http.StatusInternalServerError}, err
	}

	response := awsResponse{
		StatusCode: lambdaRequest.StatusCode,
		Body:       lambdaRequest.NewResponse(),
	}
	if ss.S.Config().IsExtraLogEnabled() {
		lambdaRequest.Log().Debug(
			"Response status code: %d. Dump: %s",
			response.StatusCode,
			response.Body)
	}
	return response, nil
}

func (service service) validateRequest(request awsRequest) error {
	// The temporary solution to be sure that AWS already validated the request,
	// see for details: https://buzzplace.atlassian.net/browse/Buzz-85
	contentType, has := request.Headers["content-type"]
	if !has {
		if contentType, has = request.Headers["Content-Type"]; !has {
			return fmt.Errorf("content type is not set")
		}
	}
	if contentType != "application/json; charset=utf-8" {
		return fmt.Errorf("invalid content type %q", contentType)
	}
	return nil
}
