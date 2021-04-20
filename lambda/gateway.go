// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package lambda

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/apigatewaymanagementapi"
	"github.com/palchukovsky/ss"
)

// Gateway describes the interface of an output gateway.
type Gateway interface {
	Send(connection ConnectionID, data interface{}) (bool, error)
	SendSerialized(connection ConnectionID, data []byte) (bool, error)
	Serialize(interface{}) ([]byte, error)
}

// NewGateway creates new gateway instance.
func NewGateway() Gateway {
	config := ss.S.Config()
	session, err := session.NewSession(
		&aws.Config{
			Region: aws.String(config.AWS.Region),
			Credentials: credentials.NewStaticCredentials(
				config.AWS.AccessKey.ID,
				config.AWS.AccessKey.Secret,
				"",
			),
		},
	)
	if err != nil {
		ss.S.Log().Panic(`Failed to create lambda session: "%v".`, err)
	}
	return gateway{client: apigatewaymanagementapi.New(session)}
}

type gateway struct {
	client *apigatewaymanagementapi.ApiGatewayManagementApi
}

func (gateway gateway) Send(connection ConnectionID, data interface{},
) (bool, error) {
	serializeData, err := gateway.Serialize(data)
	if err != nil {
		return false, err
	}
	return gateway.SendSerialized(connection, serializeData)
}

func (gateway gateway) SendSerialized(connection ConnectionID, data []byte,
) (bool, error) {
	request, _ := gateway.
		client.
		PostToConnectionRequest(
			&apigatewaymanagementapi.PostToConnectionInput{
				ConnectionId: aws.String(string(connection)),
				Data:         data,
			})
	if err := request.Send(); err != nil {
		var goneErr *apigatewaymanagementapi.GoneException
		if errors.As(err, &goneErr) {
			return false, nil
		}
		return false, fmt.Errorf(`failed to send: "%w"`, err)
	}
	return true, nil
}

func (gateway gateway) Serialize(data interface{}) ([]byte, error) {
	result, err := json.Marshal(struct {
		Method string      `json:"m"`
		Data   interface{} `json:"d"`
	}{
		Method: ss.S.Name(),
		Data:   data,
	})
	if err != nil {
		err = fmt.Errorf(`failed to serialize data: "%w"`, err)
	}
	return result, err
}
