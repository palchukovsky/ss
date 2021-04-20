// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package gatewayinstall

import (
	"context"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/apigatewayv2"
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
	CreateRoute(name string) error
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

func (client gatewayClient) CreateRoute(name string) error {
	input := apigatewayv2.CreateRouteInput{
		ApiId:    client.id,
		RouteKey: aws.String(name),
	}
	_, err := client.client.CreateRoute(context.TODO(), &input)
	return err
}

////////////////////////////////////////////////////////////////////////////////
