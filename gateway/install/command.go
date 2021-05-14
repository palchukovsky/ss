// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package gatewayinstall

import (
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/palchukovsky/ss"
)

type Command interface {
	GetName() string
	Log() ss.LogStream

	Create(GatewayClient) error
}

////////////////////////////////////////////////////////////////////////////////

type command struct {
	name string
	log  ss.LogStream
}

func newCommand(
	name string,
	log ss.LogStream,
) (command, error) {
	result := command{name: name}
	result.log = log.NewSession(
		ss.NewLogPrefix().AddVal("gatewayCommand", result.name))
	return result, nil
}

func (command command) GetName() string   { return command.name }
func (command command) Log() ss.LogStream { return command.log }

////////////////////////////////////////////////////////////////////////////////

// NewRESTCommand creates command implementation to work with REST-gateways.
func NewRESTCommand(
	name string,
	path string,
	log ss.LogStream,
) (Command, error) {
	command, err := newCommand(name, log)
	if err != nil {
		return nil, err
	}
	return restCommand{command: command}, nil
}

type restCommand struct{ command }

func (restCommand) Create(client GatewayClient) error {
	return errors.New("not implemented")
}

///////////////////////////////////////////////////////////////////////////////

//NewWSCommand creates command implementation to work with websockets gateways.
func NewWSCommand(
	name string,
	path string,
	log ss.LogStream,
) (Command, error) {

	name = strings.ReplaceAll(name, "_", " ")
	name = strings.Title(name)
	name = strings.ReplaceAll(name, " ", "")

	command, err := newCommand(name, log)

	if err != nil {
		return nil, err
	}
	return wsCommand{
			command: command,
			path:    path,
		},
		nil
}

type wsCommand struct {
	command
	path string
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

	modelFile := command.path + "/model.json"
	schema, err := ioutil.ReadFile(modelFile)
	if err != nil {
		return "",
			fmt.Errorf(`failed to read model %q schema from %q: "%w"`,
				command.name,
				modelFile,
				err)
	}
	if len(schema) == 0 {
		return "",
			fmt.Errorf(`model %q schema from %q is empty`,
				command.name,
				modelFile)
	}

	model, err := client.CreateModel(command.name, string(schema))
	if err != nil {
		return "", fmt.Errorf(`failed to create model %q: "%w"`, command.name, err)
	}

	return model, nil
}

////////////////////////////////////////////////////////////////////////////////

func newWSConnectCommand(log ss.LogStream) (Command, error) {
	command, err := newCommand("$connect", log)
	if err != nil {
		return nil, err
	}
	return wsConnectCommand{command: command}, nil
}

type wsConnectCommand struct{ command }

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

func newWSDesconnectCommand(log ss.LogStream) (Command, error) {
	command, err := newCommand("$disconnect", log)
	if err != nil {
		return nil, err
	}
	return wsDesconnectCommand{command: command}, nil
}

type wsDesconnectCommand struct{ command }

func (command wsDesconnectCommand) Create(client GatewayClient) error {
	return command.createRoute(client)
}

func (command wsDesconnectCommand) createRoute(client GatewayClient) error {
	_, err := client.CreateRoute(command.name, "Disconnect", nil, nil)
	return err
}

////////////////////////////////////////////////////////////////////////////////
