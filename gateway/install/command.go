// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package gatewayinstall

import (
	"fmt"

	"github.com/palchukovsky/ss"
	connectionupdatelambda "github.com/palchukovsky/ss/api/gateway/app/lambda/connection/update/install"
)

type command interface {
	GetName() string
	Log() ss.LogStream

	Create(GatewayClient) error
}

////////////////////////////////////////////////////////////////////////////////

type abstractCommand struct {
	name string
	log  ss.LogStream
}

func newCommand(name string, log ss.LogSession) (abstractCommand, error) {
	result := abstractCommand{name: name}
	result.log = log.NewSession(
		ss.
			NewLogPrefix(func() []ss.LogMsgAttr { return nil }).
			AddVal("gatewayCommand", result.name))
	return result, nil
}

func (command abstractCommand) GetName() string   { return command.name }
func (command abstractCommand) Log() ss.LogStream { return command.log }

///////////////////////////////////////////////////////////////////////////////

func newWSCommand(
	name string,
	schemaBuilder modelSchemaBuilder,
	log ss.LogSession,
) (command, error) {
	command, err := newCommand(name, log)
	if err != nil {
		return nil, err
	}
	return wsCommand{
			abstractCommand: command,
			schemaBuilder:   schemaBuilder,
		},
		nil
}

type wsCommand struct {
	abstractCommand
	schemaBuilder modelSchemaBuilder
}

func (command wsCommand) Create(client GatewayClient) error {
	model, err := command.createModel(client)
	if err != nil {
		return err
	}
	return command.createRoute(client, model)
}

func (command wsCommand) createRoute(
	client GatewayClient,
	model GatewayModel,
) error {
	route, err := client.CreateRoute(command.name, command.name, &model, nil)
	if err != nil {
		return err
	}
	return client.CreateRouteResponse(route)
}

func (command wsCommand) createModel(
	client GatewayClient,
) (GatewayModel, error) {

	schema, err := command.schemaBuilder.Build()
	if err != nil {
		return "",
			fmt.Errorf(
				`failed to create model schema for command %q: "%w"`,
				command.name,
				err)
	}

	model, err := client.CreateModel(command.name, schema)
	if err != nil {
		return "", fmt.Errorf(`failed to create model %q: "%w"`, command.name, err)
	}

	return model, nil
}

////////////////////////////////////////////////////////////////////////////////

func newWSConnectCommand(log ss.LogSession) (command, error) {
	command, err := newCommand("$connect", log)
	if err != nil {
		return nil, err
	}
	return wsConnectCommand{abstractCommand: command}, nil
}

type wsConnectCommand struct{ abstractCommand }

func (command wsConnectCommand) Create(client GatewayClient) error {
	return command.createRoute(client)
}

func (command wsConnectCommand) createRoute(client GatewayClient) error {
	authorizer, err := client.CreateAuthorizer("Authorizer")
	if err != nil {
		return err
	}
	_, err = client.CreateRoute(command.name, "Connect", nil, &authorizer)
	return err
}

////////////////////////////////////////////////////////////////////////////////

func newWSDesconnectCommand(log ss.LogSession) (command, error) {
	command, err := newCommand("$disconnect", log)
	if err != nil {
		return nil, err
	}
	return wsDesconnectCommand{abstractCommand: command}, nil
}

type wsDesconnectCommand struct{ abstractCommand }

func (command wsDesconnectCommand) Create(client GatewayClient) error {
	return command.createRoute(client)
}

func (command wsDesconnectCommand) createRoute(client GatewayClient) error {
	_, err := client.CreateRoute(command.name, "Disconnect", nil, nil)
	return err
}

////////////////////////////////////////////////////////////////////////////////

func newWSConnectionUpdateCommand(
	sourcePath string,
	log ss.LogSession,
) (command, error) {
	return newWSCommand(
		"ConnectionUpdate",
		newModelSchemaFromEmbedFS(connectionupdatelambda.ModelFS),
		log)
}

////////////////////////////////////////////////////////////////////////////////
