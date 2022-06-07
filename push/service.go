// Copyright 2021-2022, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package push

import (
	"context"
	"encoding/json"
	"strings"

	"firebase.google.com/go/messaging"
	"github.com/palchukovsky/ss"
	"github.com/palchukovsky/ss/crypto/aes"
	"github.com/palchukovsky/ss/db"
	"github.com/palchukovsky/ss/ddb"
	"github.com/palchukovsky/ss/push/lib"
)

////////////////////////////////////////////////////////////////////////////////

const taskQueue = 10

func NewService(db ddb.Client) *Service {
	result := &Service{
		db:           db,
		taskChan:     make(chan *task, taskQueue),
		taskSyncChan: make(chan struct{}),
		taskQueue:    taskQueue,
	}
	go func() {
		defer func() { ss.S.Log().CheckExit(recover()) }()
		result.execTasks()
	}()
	return result
}

type Service struct {
	ss.NoCopyImpl

	db     ddb.Client
	client *messaging.Client

	taskChan     chan *task
	taskSyncChan chan struct{}
	taskQueue    uint

	successCount      uint
	expiredTokenCount uint
}

func (service *Service) Sync(log ss.LogStream) {
	service.taskChan <- nil
	<-service.taskSyncChan

	queue := service.successCount + service.expiredTokenCount
	if queue == 0 {
		return
	}

	logMsg := ss.NewLogMsg(
		"%d sent, %d removed expired tokens",
		service.successCount,
		service.expiredTokenCount)
	if queue > (taskQueue * 2) {
		// Time to think about queues size increasing or about dynamic queue size.
		log.Warn(logMsg)
	} else {
		log.Debug(logMsg)
	}

	service.successCount = 0
	service.expiredTokenCount = 0
}

func (service *Service) Push(
	user ss.UserID,
	newMessage func() Message,
	log ss.LogStream,
) {

	var device lib.DeviceUserIndex
	it := service.
		db.
		Index(&device).
		Query("user = :u", ddb.Values{":u": user}).
		RequestPaged()

	var push *push
	for it.Next() {
		if push == nil {
			push = service.newPush(newMessage, log)
		}
		service.taskChan <- &task{
			Push:   push,
			Device: device,
		}
	}
}

func (service *Service) newPush(
	newMessage func() Message,
	log ss.LogStream,
) *push {
	messageSource := newMessage()

	if ss.S.Config().IsExtraLogEnabled() {
		log.Debug(ss.NewLogMsg("push message").AddDump(messageSource))
	}

	message, err := json.Marshal(map[EntityTypeName]interface{}{
		messageSource.GetType(): messageSource.GetData(),
	})
	if err != nil {
		// Can't add events into error info as it cloud not be serialized.
		log.Panic(ss.NewLogMsg("failed to serialize push message").AddErr(err))
	}

	if service.client == nil {
		service.client, err = ss.S.Firebase().Messaging(context.Background())
		if err != nil {
			log.Panic(ss.NewLogMsg("failed to get Firebase Messaging").AddErr(err))
		}
	}

	return newPush(message, service, log)
}

func (service *Service) execTasks() {
	for {
		task := <-service.taskChan
		if task == nil {
			service.taskSyncChan <- struct{}{}
			continue
		}
		task.Push.send(task.Device)
	}
}

////////////////////////////////////////////////////////////////////////////////

type task struct {
	Push   *push
	Device lib.DeviceUserIndex
}

////////////////////////////////////////////////////////////////////////////////

func newPush(message []byte, service *Service, log ss.LogStream) *push {
	return &push{
		service: service,
		log:     log,
		message: message,
	}
}

type push struct {
	ss.NoCopyImpl

	service *Service
	log     ss.LogStream

	message []byte
}

func (push *push) send(device lib.DeviceUserIndex) {

	message := &messaging.Message{
		Data:  push.newMessage(device.Key),
		Token: string(device.FCMToken),
		Android: &messaging.AndroidConfig{
			Priority: "high",
		},
		APNS: &messaging.APNSConfig{ // for iOS
			Payload: &messaging.APNSPayload{
				Aps: &messaging.Aps{
					ContentAvailable: true, // also "apns-priority" header
					Alert: &messaging.ApsAlert{
						Body: "",
					},
				},
			},
			Headers: map[string]string{
				"apns-push-type": "background",
				"apns-priority":  "5", // must be `5` when `contentAvailable` is true.
				"apns-topic":     ss.S.Config().App.IOS.Bundle,
			},
		},
	}

	_, err := push.service.client.Send(context.Background(), message)
	if err == nil {
		push.service.successCount += 1

		if ss.S.Config().IsExtraLogEnabled() {
			push.log.Debug(
				ss.
					NewLogMsg("push message sent").
					AddVal("fcm", device.FCMToken).
					AddDump(string(push.message)))
		}
		return
	}

	if strings.Contains(err.Error(), "registration-token-not-registered") {
		push.service.expiredTokenCount += 1
		push.deleteExpiredToken(device)
		return
	}

	push.log.Panic(
		ss.
			NewLogMsg("failed to send push messages").
			AddErr(err).
			AddDump(message))

}

func (push *push) newMessage(key db.DeviceCryptoKey) map[string]string {
	data, auth := aes.Encrypt(push.message, key)
	return map[string]string{
		messageDataFieldName: data,
		messageAuthFieldName: auth,
	}
}

func (push *push) deleteExpiredToken(device lib.DeviceUserIndex) {
	isSuccess := push.
		service.
		db.
		DeleteIfExisting(db.NewDeviceKey(device.FCMToken)).
		Request().
		IsSuccess()
	if !isSuccess {
		push.log.Debug(ss.NewLogMsg("expired token already changed"))
		return
	}
	push.log.Debug(ss.NewLogMsg("deleted expired token"))
}

////////////////////////////////////////////////////////////////////////////////
