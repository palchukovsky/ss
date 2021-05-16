// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package gatewayinstall

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/palchukovsky/ss"
)

type Gateway interface {
	GetName() string
	Log() ss.LogStream

	Create(Client) error
	Delete(Client) error
	Deploy(Client) error
}

// NewGateway creates new gateway instance.
func NewAppGateway(log ss.LogStream) Gateway {
	return NewGateway(
		ss.S.Config().AWS.Gateway.App.ID,
		"app",
		NewGatewayCommadsReader(NewWSCommand),
		log,
	)
}

// NewGateway creates new gateway instance.
func NewGateway(
	id string,
	name string,
	reader GatewayCommadsReader,
	log ss.LogStream,
) Gateway {

	log = log.NewSession(ss.NewLogPrefix().AddVal("gateway", name))

	commands, err := reader.Read(name, log)
	if err != nil {
		log.Panic(
			ss.
				NewLogMsg(`failed to read API commands directory to build gateway`).
				AddErr(err))
	}

	{
		connect, err := newWSConnectCommand(log)
		if err != nil {
			log.Panic(
				ss.NewLogMsg(`failed to create API command "connect"`).AddErr(err))
		}
		commands = append(commands, connect)
	}
	{
		disconnect, err := newWSDesconnectCommand(log)
		if err != nil {
			log.Panic(
				ss.NewLogMsg(`failed to create API command "discconnect"`).AddErr(err))
		}
		commands = append(commands, disconnect)
	}

	return gateway{
		id:       id,
		name:     name,
		log:      log,
		commands: commands,
	}
}

////////////////////////////////////////////////////////////////////////////////

type GatewayCommadsReader interface {
	Read(name string, log ss.LogStream) ([]Command, error)
}

func NewGatewayCommadsReader(
	newCommand func(name, path string, log ss.LogStream) (Command, error),
) GatewayCommadsReader {
	return gatewayCommadsReader{newCommand: newCommand}
}

type gatewayCommadsReader struct {
	newCommand func(name, path string, log ss.LogStream) (Command, error)
}

func (reader gatewayCommadsReader) Read(
	name string,
	log ss.LogStream,
) ([]Command, error) {
	result := []Command{}

	err := filepath.Walk(
		"cmd/lambda/api/"+name,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			command, err := reader.newCommand(info.Name(), path, log)
			if err != nil {
				return fmt.Errorf(`failed to create commad %q: "%w"`, path, err)
			}
			command.Log().Info(ss.NewLogMsg("found command by path %q", path))

			result = append(result, command)
			return nil
		})

	if err != nil {
		log.Panic(
			ss.NewLogMsg(
				`failed to read API commands directory to build gateway %q`,
				name).
				AddErr(err))
	}

	return result, nil
}

////////////////////////////////////////////////////////////////////////////////

type gateway struct {
	id       string
	name     string
	log      ss.LogStream
	commands []Command
}

func (gateway gateway) GetName() string   { return gateway.name }
func (gateway gateway) Log() ss.LogStream { return gateway.log }

func (gateway gateway) Create(client Client) error {
	gatewayClient := client.NewGatewayClient(gateway.id)
	for _, commad := range gateway.commands {
		if err := commad.Create(gatewayClient); err != nil {
			return fmt.Errorf(`failed to create commad %q/%q: "%w"`,
				gateway.name,
				commad.GetName(),
				err)
		}
	}
	return nil
}

func (gateway gateway) Delete(client Client) error {
	gatewayClient := client.NewGatewayClient(gateway.id)
	if err := gatewayClient.DeleteRoutes(); err != nil {
		return err
	}
	if err := gatewayClient.DeleteAuthorizers(); err != nil {
		return err
	}
	if err := gatewayClient.DeleteModels(); err != nil {
		return err
	}
	return nil
}

func (gateway gateway) Deploy(client Client) error {
	return client.NewGatewayClient(gateway.id).Deploy()
}
