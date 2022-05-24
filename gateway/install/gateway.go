// Copyright 2021-2022, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package gatewayinstall

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"unicode"

	"github.com/palchukovsky/ss"
)

type Gateway interface {
	GetName() string
	Log() ss.LogStream

	Create(sourcePath string, client Client) error
	Delete(Client) error
	Deploy(Client) error
}

// NewGateway creates new gateway instance.
func NewGateway(
	id string,
	name string,
	isUserCanRemoveHimself bool,
	log ss.LogSession,
) Gateway {
	return gateway{
		id:   id,
		name: name,
		log: log.NewSession(
			func() ss.LogPrefix {
				return ss.
					NewLogPrefix(func() []ss.LogMsgAttr { return nil }).
					AddVal("gateway", name)
			}),
		isUserCanRemoveHimself: isUserCanRemoveHimself,
	}
}

////////////////////////////////////////////////////////////////////////////////

type GatewayCommadsReader interface {
	Read(name string, log ss.LogSession) ([]command, error)
}

func newGatewayCommandsReader(
	sourcePath string,
	newCommand func(name, path string, log ss.LogSession) (command, error),
) GatewayCommadsReader {
	return gatewayCommadsReader{
		sourcePath: sourcePath,
		newCommand: newCommand,
	}
}

type gatewayCommadsReader struct {
	sourcePath string
	newCommand func(name, path string, log ss.LogSession) (command, error)
}

func (reader gatewayCommadsReader) Read(
	name string,
	log ss.LogSession,
) ([]command, error) {
	result := []command{}

	sourcePath := reader.sourcePath
	if !strings.HasSuffix(sourcePath, "/") {
		sourcePath += "/"
	}
	sourcePath += "cmd/lambda/api/" + name + "/"

	err := filepath.Walk(
		sourcePath,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() {
				return nil
			}
			if _, err := os.Stat(path + "/model.json"); os.IsNotExist(err) {
				return nil
			}

			name := strings.TrimPrefix(path, sourcePath)
			name = strings.ReplaceAll(name, "/", " ")
			name = strings.ReplaceAll(name, "_", " ")
			name = strings.ReplaceAll(name, " ", "")
			{
				runes := []rune(name)
				runes[0] = unicode.ToUpper(runes[0])
				name = string(runes)
			}

			command, err := reader.newCommand(name, path, log)
			if err != nil {
				return fmt.Errorf(
					`failed to create command %q by path %q: "%w"`,
					name,
					path,
					err)
			}
			command.Log().Info(
				ss.NewLogMsg("found command %q by path %q", name, path))

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
	log  ss.LogSession

	isUserCanRemoveHimself bool
}

func (gateway gateway) GetName() string   { return gateway.name }
func (gateway gateway) Log() ss.LogStream { return gateway.log }

func (gateway gateway) Create(sourcePath string, client Client) error {
	gatewayClient := client.NewGatewayClient(gateway.id)
	for _, commad := range gateway.readCommads(sourcePath) {
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

func (gateway *gateway) readCommads(sourcePath string) []command {

	reader := newGatewayCommandsReader(
		sourcePath,
		func(name, path string, log ss.LogSession) (command, error) {
			return newWSCommand(name, newModelSchemaFromFileBuilder(path), log)
		})
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
				ss.NewLogMsg(`failed to create API command "disconnect"`).AddErr(err))
		}
		result = append(result, disconnect)
	}
	{
		connectionUpdate, err := newWSConnectionUpdateCommand(
			sourcePath,
			gateway.log)
		if err != nil {
			log.Panic(
				ss.
					NewLogMsg(`failed to create API command "connection update"`).
					AddErr(err))
		}
		result = append(result, connectionUpdate)
	}
	if gateway.isUserCanRemoveHimself {
		userDelete, err := newWSUserDeleteCommand(sourcePath, gateway.log)
		if err != nil {
			log.Panic(
				ss.
					NewLogMsg(`failed to create API command "user delete"`).
					AddErr(err))
		}
		result = append(result, userDelete)
	}

	return result
}
