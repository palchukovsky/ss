// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package lambda

import (
	"encoding/json"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"time"

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
		messageChan: make(chan gatewayMessage, gatewaySendSessionWarnLevel),
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

const gatewaySendSessionWarnLevel = 100

type gatewaySendSession struct {
	gateway Gateway
	log     ss.LogSession
	sync    sync.WaitGroup

	messageChan chan gatewayMessage

	added uint32
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
	session.messageChan <- gatewayMessage{}
	session.sync.Wait()

	if session.sends+session.skips > gatewaySendSessionWarnLevel {
		session.log.Warn(ss.NewLogMsg(
			"processed %d gateway messages, %d sent, %d skipped",
			session.sends+session.skips,
			session.sends,
			session.skips))
	}

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

	if atomic.AddUint32(&session.added, 1)%gatewaySendSessionWarnLevel == 0 {
		session.log.Warn(
			ss.NewLogMsg(
				"already added %d messages to send through gateway...",
				atomic.LoadUint32(&session.added)))
	}
}

func (session *gatewaySendSession) SendSerialized(
	connection ss.ConnectionID,
	data []byte,
) {
	session.Send(connection, data)
}

func (session *gatewaySendSession) runSender() {
	defer session.sync.Done()

	for {
		message := <-session.messageChan
		if len(message.Connection) == 0 {
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
			defer session.sync.Done()

			select {
			case <-doneChan:
				return
			case <-time.After(ss.LambdaMaxRunTimeInclusive):
				break
			case <-ss.S.GetLambdaTimeout():
				break
			}

			logMessage := ss.NewLogMsg("gateway message sending timeout").
				Add(message.Connection)
			if isSerialized {
				logMessage.AddDump(string(data))
			} else {
				logMessage.AddDump(message.Data)
			}
			session.log.Warn(logMessage)
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

	var processed uint32
	if err := request.Send(); err != nil {

		var goneErr *apigatewaymanagementapi.GoneException
		if errors.As(err, &goneErr) {
			logMessage := ss.NewLogMsg("no connection to send gateway message").
				Add(connection)
			if ss.S.Config().IsExtraLogEnabled() {
				logMessage.AddDump(string(data))
			}
			session.log.Debug(logMessage)
		} else {
			session.log.Error(
				ss.NewLogMsg("failed to send gateway message").
					AddErr(err).
					Add(connection).
					AddDump(string(data)))
		}

		processed = atomic.AddUint32(&session.skips, 1) +
			atomic.LoadUint32(&session.sends)

	} else {

		if ss.S.Config().IsExtraLogEnabled() {
			session.log.Debug(
				ss.NewLogMsg("sent").Add(connection).AddDump(string(data)))
		}

		processed = atomic.AddUint32(&session.sends, 1) +
			atomic.LoadUint32(&session.skips)

	}

	if processed%gatewaySendSessionWarnLevel == 0 {
		sends := atomic.LoadUint32(&session.sends)
		skips := atomic.LoadUint32(&session.skips)
		total := sends + skips
		logMessage := ss.NewLogMsg(
			"already processed %d gateway messages, %d sent, %d skipped...",
			total,
			sends,
			skips)
		if total <= gatewaySendSessionWarnLevel {
			session.log.Info(logMessage)
		} else {
			session.log.Warn(logMessage)
		}
	}
}

////////////////////////////////////////////////////////////////////////////////
