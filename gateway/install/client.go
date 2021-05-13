// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package gatewayinstall

import (
	"context"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2/types"
	"github.com/palchukovsky/ss"
)

////////////////////////////////////////////////////////////////////////////////

type Client interface {
	NewGatewayClient(id string) GatewayClient
}

// NewClient creates new client instance.
func NewClient() Client {
	return client{
		client: apigatewayv2.NewFromConfig(ss.S.NewAWSConfig()),
	}
}

type client struct{ client *apigatewayv2.Client }

func (client client) NewGatewayClient(id string) GatewayClient {
	return gatewayClient{
		client: client.client,
		id:     aws.String(id),
	}
}

////////////////////////////////////////////////////////////////////////////////

type GatewayClient interface {
	CreateModel(name, schema string) error
	CreateRoute(name, description string) error
}

type gatewayClient struct {
	client *apigatewayv2.Client
	id     *string
}

func (client gatewayClient) CreateModel(name, schema string) error {
	input := apigatewayv2.CreateModelInput{
		ApiId:       client.id,
		Name:        aws.String(name),
		ContentType: aws.String("application/json"),
		Schema:      aws.String(schema),
	}
	_, err := client.client.CreateModel(context.TODO(), &input)
	return err
}

func (client gatewayClient) CreateRoute(
	name string,
	description string,
) error {

	lambda := fmt.Sprintf(
		"arn:aws:apigateway:%s:lambda:path/2015-03-31/functions/arn:aws:lambda:%s:%s:function:${stageVariables.lambdaPrefix}%s/invocations",
		ss.S.Config().AWS.Region,
		ss.S.Config().AWS.Region,
		ss.S.Config().AWS.AccountID,
		name)
	integrationInput := apigatewayv2.CreateIntegrationInput{
		ApiId:                   client.id,
		IntegrationType:         types.IntegrationTypeAwsProxy,
		ContentHandlingStrategy: types.ContentHandlingStrategyConvertToText,
		IntegrationUri:          aws.String(lambda),
		Description:             aws.String(description),
	}
	integrationOutput, err := client.client.CreateIntegration(
		context.TODO(),
		&integrationInput)
	if err != nil {
		return err
	}

	{
		input := apigatewayv2.CreateIntegrationResponseInput{
			ApiId:                  client.id,
			IntegrationId:          integrationOutput.IntegrationId,
			IntegrationResponseKey: aws.String("$default"),
		}
		_, err := client.client.CreateIntegrationResponse(context.TODO(), &input)
		if err != nil {
			return err
		}
	}

	target := aws.String(
		"integrations/" + aws.ToString(integrationOutput.IntegrationId))
	routeInput := apigatewayv2.CreateRouteInput{
		ApiId:                            client.id,
		RouteKey:                         aws.String(name),
		ModelSelectionExpression:         aws.String("$request.body.m"),
		RouteResponseSelectionExpression: aws.String("$default"),
		RequestModels:                    map[string]string{"$default": name},
		Target:                           target,
	}
	routeOutput, err := client.client.CreateRoute(context.TODO(), &routeInput)
	if err != nil {
		return err
	}

	{
		input := apigatewayv2.CreateRouteResponseInput{
			ApiId:            client.id,
			RouteId:          routeOutput.RouteId,
			RouteResponseKey: aws.String("$default"),
		}
		_, err := client.client.CreateRouteResponse(context.TODO(), &input)
		if err != nil {
			return err
		}
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////
