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
	Log() ss.ServiceLogStream

	Create(Client) error
}

// NewGateway creates new gateway instance.
func NewGateway(
	id string,
	name string,
	reader GatewayCommadsReader,
	log ss.ServiceLogStream,
) Gateway {

	log = log.NewSession(name)

	commands, err := reader.Read(name, log)
	if err != nil {
		log.Panic(`Failed to read API commands directory to build gateway %q: "%v"`,
			name,
			err)
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
	Read(name string, log ss.ServiceLogStream) ([]Command, error)
}

func NewGatewayCommadsReader(
	newCommand func(name, path string, log ss.ServiceLogStream) (Command, error),
) GatewayCommadsReader {
	return gatewayCommadsReader{newCommand: newCommand}
}

type gatewayCommadsReader struct {
	newCommand func(name, path string, log ss.ServiceLogStream) (Command, error)
}

func (reader gatewayCommadsReader) Read(
	name string,
	log ss.ServiceLogStream,
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
			command.Log().Info("Found command by path %q.", path)

			result = append(result, command)
			return nil
		})

	if err != nil {
		log.Panic(`Failed to read API commands directory to build gateway %q: "%v"`,
			name,
			err)
	}

	return result, nil
}

////////////////////////////////////////////////////////////////////////////////

type gateway struct {
	id       string
	name     string
	log      ss.ServiceLogStream
	commands []Command
}

func (gateway gateway) GetName() string          { return gateway.name }
func (gateway gateway) Log() ss.ServiceLogStream { return gateway.log }

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
