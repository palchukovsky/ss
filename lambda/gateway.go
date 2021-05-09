// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package lambda

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/apigatewaymanagementapi"
	"github.com/palchukovsky/ss"
)

////////////////////////////////////////////////////////////////////////////////

// Gateway describes the interface of an output gateway.
type Gateway struct {
	client *apigatewaymanagementapi.ApiGatewayManagementApi
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
			Endpoint: aws.String(config.AWS.Gateway.App.Endpoint),
		},
	)
	if err != nil {
		ss.S.Log().Panic(
			ss.NewLogMsg(`failed to create lambda session`).AddErr(err))
	}
	return Gateway{client: apigatewaymanagementapi.New(session)}
}

// NewSessionGatewaySendSession creates a new session to send data thought
// gateway which has to be closed by method Close after usage.
func (gateway Gateway) NewSessionGatewaySendSession(
	log ss.LogSession,
) *gatewaySendSession {
	result := gatewaySendSession{
		gateway:     gateway,
		messageChan: make(chan gatewayMessage),
		log:         log,
	}
	result.sync.Add(1)
	go result.runSender()
	return &result
}

func (gateway Gateway) Serialize(data interface{}) ([]byte, error) {
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

////////////////////////////////////////////////////////////////////////////////

type gatewaySendSession struct {
	gateway Gateway
	log     ss.LogSession
	sync    sync.WaitGroup

	messageChan chan gatewayMessage

	sends uint32
	skips uint32
}

type gatewayMessage struct {
	Connection ss.ConnectionID
	Data       interface{}
}

// CloseAndGetStat closes the session and returns number of sent messages
// and number of skipped (by errors or disconnection).
func (session *gatewaySendSession) CloseAndGetStat() (uint32, uint32) {
	close(session.messageChan)
	session.sync.Wait()
	return session.sends, session.skips
}

func (session *gatewaySendSession) Send(
	connection ss.ConnectionID,
	data interface{},
) {
	session.messageChan <- gatewayMessage{
		Connection: connection,
		Data:       data,
	}
}

func (session *gatewaySendSession) SendSerialized(
	connection ss.ConnectionID,
	data []byte,
) {
	session.messageChan <- gatewayMessage{
		Connection: connection,
		Data:       data,
	}
}

func (session *gatewaySendSession) runSender() {
	defer session.sync.Done()

	for {
		message, isOpen := <-session.messageChan
		if !isOpen {
			break
		}

		data, isSerialized := message.Data.([]byte)
		if !isSerialized {
			var err error
			if data, err = session.gateway.Serialize(message.Data); err != nil {
				session.log.Panic(
					ss.NewLogMsg("failed to serialize message for gateway").
						AddErr(err).
						Add(message.Connection).
						AddDump(data))
			}
		}

		doneChan := make(chan struct{})
		go func() {
			session.send(message.Connection, data)
			doneChan <- struct{}{}
		}()
		session.sync.Add(1)
		go func() {
			select {
			case <-doneChan:
				break
			case <-ss.S.GetLambdaTimeout():
				logMessage := ss.NewLogMsg("gateway message sending timeout").
					Add(message.Connection)
				if isSerialized {
					logMessage.AddDump(string(data))
				} else {
					logMessage.AddDump(message.Data)
				}
				session.log.Warn(logMessage)
			}
			session.sync.Done()
		}()

	}
}

func (session *gatewaySendSession) send(
	connection ss.ConnectionID,
	data []byte,
) {

	request, _ := session.
		gateway.
		client.
		PostToConnectionRequest(
			&apigatewaymanagementapi.PostToConnectionInput{
				ConnectionId: aws.String(string(connection)),
				Data:         data,
			})

	if err := request.Send(); err != nil {
		atomic.AddUint32(&session.skips, 1)

		var goneErr *apigatewaymanagementapi.GoneException
		if errors.As(err, &goneErr) {
			logMessage := ss.NewLogMsg("no connection to send gateway message").
				Add(connection)
			if ss.S.Config().IsExtraLogEnabled() {
				logMessage.AddDump(string(data))
			}
			session.log.Debug(logMessage)
			return
		}

		session.log.Error(
			ss.NewLogMsg("failed to send gateway message").
				AddErr(err).
				Add(connection).
				AddDump(string(data)))

		return
	}

	if ss.S.Config().IsExtraLogEnabled() {
		session.log.Debug(
			ss.NewLogMsg("sent").Add(connection).AddDump(string(data)))
	}
	atomic.AddUint32(&session.sends, 1)
}

////////////////////////////////////////////////////////////////////////////////
