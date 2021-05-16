// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package gatewayinstall

import (
	"fmt"
	"log"
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
		log)
}

// NewGateway creates new gateway instance.
func NewGateway(
	id string,
	name string,
	log ss.LogStream,
) Gateway {
	return gateway{
		id:   id,
		name: name,
		log:  log.NewSession(ss.NewLogPrefix().AddVal("gateway", name)),
	}
}

////////////////////////////////////////////////////////////////////////////////

type GatewayCommadsReader interface {
	Read(name string, log ss.LogStream) ([]command, error)
}

func newGatewayCommandsReader(
	newCommand func(name, path string, log ss.LogStream) (command, error),
) GatewayCommadsReader {
	return gatewayCommadsReader{newCommand: newCommand}
}

type gatewayCommadsReader struct {
	newCommand func(name, path string, log ss.LogStream) (command, error)
}

func (reader gatewayCommadsReader) Read(
	name string,
	log ss.LogStream,
) ([]command, error) {
	result := []command{}

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
	id   string
	name string
	log  ss.LogStream
}

func (gateway gateway) GetName() string   { return gateway.name }
func (gateway gateway) Log() ss.LogStream { return gateway.log }

func (gateway gateway) Create(client Client) error {
	gatewayClient := client.NewGatewayClient(gateway.id)
	for _, commad := range gateway.readCommads() {
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
func (gateway *gateway) readCommads() []command {
	reader := newGatewayCommandsReader(newWSCommand)

	result, err := reader.Read(gateway.name, gateway.log)
	if err != nil {
		log.Panic(
			ss.
				NewLogMsg(`failed to read API commands directory to build gateway`).
				AddErr(err))
	}

	{
		connect, err := newWSConnectCommand(gateway.log)
		if err != nil {
			log.Panic(
				ss.NewLogMsg(`failed to create API command "connect"`).AddErr(err))
		}
		result = append(result, connect)
	}
	{
		disconnect, err := newWSDesconnectCommand(gateway.log)
		if err != nil {
			log.Panic(
				ss.NewLogMsg(`failed to create API command "discconnect"`).AddErr(err))
		}
		result = append(result, disconnect)
	}

	return result
}
