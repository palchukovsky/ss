// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package wsgatewaylambda

import (
	"encoding/json"
	"fmt"
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

	lambdaRequest, err := newRequest(request, service.Gateway)
	if err != nil {
		return awsResponse{}, err
	}
	defer func() { lambdaRequest.Log().CheckExit(recover()) }()

	if ss.S.Config().IsExtraLogEnabled() {
		lambdaRequest.Log().Debug(
			"Lambda request dump: %s.",
			ss.Dump(lambdaRequest.AWSRequest))
	}

	var responseBody string
	if err := service.lambda.Execute(&lambdaRequest); err != nil {
		lambdaRequest.Log().Error(`Lambda execution error: "%v". Request dump: %s.`,
			err, ss.Dump(lambdaRequest.AWSRequest))
		responseBody = newErrorResponseBody(lambdaRequest)
	} else {
		responseBody = newSuccessResponseBody(lambdaRequest)
	}

	if ss.S.Config().IsExtraLogEnabled() {
		lambdaRequest.Log().Debug("Response dump: %s.", responseBody)
	}
	return awsResponse{StatusCode: http.StatusOK, Body: responseBody}, nil
}

func newSuccessResponseBody(request request) string {
	id := request.ReadRemoteID()
	if id == nil {
		return ""
	}
	responseDump, err := json.Marshal(struct {
		ID   string      `json:"i"`
		Data interface{} `json:"d"`
	}{
		ID:   *id,
		Data: request.ResponseBody,
	})
	if err != nil {
		ss.S.Log().Panic(`Failed to marshal response: "%w". Dump: %s`,
			err, ss.Dump(request.ResponseBody))
	}
	return string(responseDump)
}

func newErrorResponseBody(request request) string {
	id := request.ReadRemoteID()
	if id == nil {
		return ""
	}
	return fmt.Sprintf(`{"i":%q,"e":"server error"}`, *id)
}
