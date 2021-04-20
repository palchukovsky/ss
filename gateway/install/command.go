// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package gatewayinstall

import (
	"fmt"
	"io/ioutil"
	"strings"

	"github.com/palchukovsky/ss"
)

type Command interface {
	GetName() string
	Log() ss.ServiceLog

	Create(GatewayClient) error
}

////////////////////////////////////////////////////////////////////////////////

type command struct {
	name string
	log  ss.ServiceLog
	path string
}

func newCommand(
	name string,
	path string,
	log ss.ServiceLog,
) (command, error) {

	result := command{
		name: name,
		path: path,
	}

	result.name = strings.ReplaceAll(result.name, "_", " ")
	result.name = strings.Title(result.name)
	result.name = strings.ReplaceAll(result.name, " ", "")

	result.log = log.NewSession(result.name)

	return result, nil
}

func (command command) GetName() string    { return command.name }
func (command command) Log() ss.ServiceLog { return command.log }

func (command command) createModel(client GatewayClient) error {

	modelFile := command.path + "/model.json"
	schema, err := ioutil.ReadFile(modelFile)
	if err != nil {
		return fmt.Errorf(`failed to read model %q schema from %q: "%w"`,
			command.name,
			modelFile,
			err)
	}
	if len(schema) == 0 {
		return fmt.Errorf(`model %q schema from %q is empty`,
			command.name,
			modelFile)
	}

	err = client.CreateModel(command.name, string(schema))
	if err != nil {
		return fmt.Errorf(`failed to create model %q: "%w"`, command.name, err)
	}

	return nil
}

////////////////////////////////////////////////////////////////////////////////

// NewRESTCommand creates command implementation to work with REST-gateways.
func NewRESTCommand(
	name string,
	path string,
	log ss.ServiceLog,
) (Command, error) {
	command, err := newCommand(name, path, log)
	if err != nil {
		return nil, err
	}
	return restCommand{command: command}, nil
}

type restCommand struct{ command }

func (command restCommand) Create(client GatewayClient) error {
	if err := command.createModel(client); err != nil {
		return err
	}
	return command.createRoute(client)
}

func (command restCommand) createRoute(client GatewayClient) error {
	return client.CreateRoute(command.name)
}

////////////////////////////////////////////////////////////////////////////////

// NewWSCommand creates command implementation to work with websockets gateways.
func NewWSCommand(
	name string,
	path string,
	log ss.ServiceLog,
) (Command, error) {
	command, err := newCommand(name, path, log)
	if err != nil {
		return nil, err
	}
	return wsCommand{command: command}, nil
}

type wsCommand struct{ command }

func (command wsCommand) Create(GatewayClient) error { return nil }

////////////////////////////////////////////////////////////////////////////////
