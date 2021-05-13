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

type GatewayModel string

type GatewayAuthorizer string

type GatewayRoute struct {
	Route        string
	Intergartion string
}

type GatewayClient interface {
	CreateModel(name, schema string) (GatewayModel, error)
	CreateRoute(
		name string,
		lambda string,
		model *GatewayModel,
		auth *GatewayAuthorizer,
	) (GatewayRoute, error)
	CreateRouteResponse(GatewayRoute) error
	CreateAuthorizer(name string) (GatewayAuthorizer, error)
}

type gatewayClient struct {
	client *apigatewayv2.Client
	id     *string
}

func (client gatewayClient) CreateModel(
	name string,
	schema string,
) (GatewayModel, error) {
	input := apigatewayv2.CreateModelInput{
		ApiId:       client.id,
		Name:        aws.String(name),
		ContentType: aws.String("application/json"),
		Schema:      aws.String(schema),
	}
	if _, err := client.client.CreateModel(context.TODO(), &input); err != nil {
		return "", err
	}
	return GatewayModel(name), nil
}

func (client gatewayClient) CreateRoute(
	name string,
	lambda string,
	model *GatewayModel,
	auth *GatewayAuthorizer,
) (GatewayRoute, error) {

	uri := fmt.Sprintf(
		"arn:aws:apigateway:%s:lambda:path/2015-03-31/functions/arn:aws:lambda:%s:%s:function:${stageVariables.lambdaPrefix}%s/invocations",
		ss.S.Config().AWS.Region,
		ss.S.Config().AWS.Region,
		ss.S.Config().AWS.AccountID,
		lambda)
	integrationInput := apigatewayv2.CreateIntegrationInput{
		ApiId:                   client.id,
		IntegrationType:         types.IntegrationTypeAwsProxy,
		ContentHandlingStrategy: types.ContentHandlingStrategyConvertToText,
		IntegrationUri:          aws.String(uri),
	}
	integrationOutput, err := client.client.CreateIntegration(
		context.TODO(),
		&integrationInput)
	if err != nil {
		return GatewayRoute{}, err
	}

	target := aws.String(
		"integrations/" + aws.ToString(integrationOutput.IntegrationId))
	routeInput := apigatewayv2.CreateRouteInput{
		ApiId:                            client.id,
		RouteKey:                         aws.String(name),
		RouteResponseSelectionExpression: aws.String("$default"),
		Target:                           target,
	}
	if model != nil {
		routeInput.ModelSelectionExpression = aws.String("$request.body.m")
		routeInput.RequestModels = map[string]string{"$default": lambda}
	}
	if auth != nil {
		routeInput.AuthorizationType = types.AuthorizationTypeCustom
		routeInput.AuthorizerId = aws.String(string(*auth))
	}
	routeOutput, err := client.client.CreateRoute(context.TODO(), &routeInput)
	if err != nil {
		return GatewayRoute{}, err
	}

	return GatewayRoute{
			Route:        aws.ToString(routeOutput.RouteId),
			Intergartion: aws.ToString(integrationOutput.IntegrationId),
		},
		nil

}

func (client gatewayClient) CreateRouteResponse(route GatewayRoute) error {

	{
		input := apigatewayv2.CreateIntegrationResponseInput{
			ApiId:                  client.id,
			IntegrationId:          aws.String(route.Intergartion),
			IntegrationResponseKey: aws.String("$default"),
		}
		_, err := client.client.CreateIntegrationResponse(context.TODO(), &input)
		if err != nil {
			return err
		}
	}

	{
		input := apigatewayv2.CreateRouteResponseInput{
			ApiId:            client.id,
			RouteId:          aws.String(route.Route),
			RouteResponseKey: aws.String("$default"),
		}
		_, err := client.client.CreateRouteResponse(context.TODO(), &input)
		if err != nil {
			return err
		}
	}

	return nil
}

func (client gatewayClient) CreateAuthorizer(
	name string,
) (GatewayAuthorizer, error) {
	config := ss.S.Config().AWS
	input := apigatewayv2.CreateAuthorizerInput{
		ApiId:          client.id,
		AuthorizerType: types.AuthorizerTypeRequest,
		IdentitySource: []string{"route.request.header.Auth"},
		Name:           aws.String("Authorizer"),
		AuthorizerUri: aws.String(
			fmt.Sprintf(
				"arn:aws:apigateway:%s:lambda:path/2015-03-31/functions/arn:aws:lambda:%s:%s:function:${stageVariables.lambdaPrefix}%s/invocations",
				config.Region,
				config.Region,
				config.AccountID,
				name,
			)),
	}
	output, err := client.client.CreateAuthorizer(context.TODO(), &input)
	if err != nil {
		return "", err
	}
	return GatewayAuthorizer(aws.ToString(output.AuthorizerId)), nil
}

////////////////////////////////////////////////////////////////////////////////
