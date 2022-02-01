// Copyright 2021-2022, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package usercontentupdatelambda

import (
	"sync"

	"github.com/aws/aws-lambda-go/events"
	"github.com/palchukovsky/ss"
	apidbevent "github.com/palchukovsky/ss/api/dbevent"
	"github.com/palchukovsky/ss/ddb"
	dbeventlambda "github.com/palchukovsky/ss/lambda/dbevent"
)

type lambda struct {
	db ddb.Client

	newNewUserUpdater     *UpdaterFactory
	newUpdatedUserUpdater *UpdaterFactory
	newDeletedUserUpdater *UpdaterFactory
}

func newLambda(
	newNewUserUpdater *UpdaterFactory,
	newUpdatedUserUpdater *UpdaterFactory,
	newDeletedUserUpdater *UpdaterFactory,
) dbeventlambda.Lambda {
	return lambda{
		db:                    ddb.GetClientInstance(),
		newNewUserUpdater:     newNewUserUpdater,
		newUpdatedUserUpdater: newUpdatedUserUpdater,
		newDeletedUserUpdater: newDeletedUserUpdater,
	}
}

func (lambda lambda) Execute(request dbeventlambda.Request) error {
	for _, event := range request.GetEvents() {
		lambda.execute(request, event)
	}
	return nil
}

func (lambda lambda) execute(
	request dbeventlambda.Request,
	event events.DynamoDBEventRecord,
) {
	switch events.DynamoDBOperationType(event.EventName) {
	case events.DynamoDBOperationTypeInsert:
		lambda.handelInsert(request, event)
	case events.DynamoDBOperationTypeModify:
		lambda.handelModify(request, event)
	case events.DynamoDBOperationTypeRemove:
		lambda.handelRemove(request, event)
	}
}

func (lambda lambda) handelInsert(
	request dbeventlambda.Request,
	event events.DynamoDBEventRecord,
) {
	if lambda.newNewUserUpdater == nil {
		return
	}
	image := event.Change.NewImage
	if lambda.isIndexImage(image) {
		return
	}
	user := lambda.initHandling(image, event.EventName, request)
	defer func() { request.PopLogSession(recover()) }()

	lambda.run(user, request, event, *lambda.newNewUserUpdater)
}

func (lambda lambda) handelModify(
	request dbeventlambda.Request,
	event events.DynamoDBEventRecord,
) {
	if lambda.newUpdatedUserUpdater == nil {
		return
	}
	image := event.Change.NewImage
	if lambda.isIndexImage(image) {
		return
	}
	user := lambda.initHandling(image, event.EventName, request)
	defer func() { request.PopLogSession(recover()) }()

	lambda.run(user, request, event, *lambda.newUpdatedUserUpdater)
}

func (lambda lambda) handelRemove(
	request dbeventlambda.Request,
	event events.DynamoDBEventRecord,
) {
	image := event.Change.OldImage
	if lambda.isIndexImage(image) {
		return
	}
	user := lambda.initHandling(image, event.EventName, request)
	defer func() { request.PopLogSession(recover()) }()

	updaters := []UpdaterFactory{newDeleter}
	if lambda.newDeletedUserUpdater != nil {
		updaters = append(updaters, *lambda.newDeletedUserUpdater)
	}

	lambda.run(user, request, event, updaters...)
}

func (lambda) isIndexImage(
	image map[string]events.DynamoDBAttributeValue,
) bool {
	if len(image) != 2 {
		return false
	}
	if _, hasID := image["id"]; !hasID {
		return false
	}
	if _, hasUser := image["user"]; !hasUser {
		return false
	}
	return true
}

func (lambda) initHandling(
	image map[string]events.DynamoDBAttributeValue,
	eventName string,
	request dbeventlambda.Request,
) ss.UserID {
	var record struct {
		User ss.UserID `json:"id"`
	}

	apidbevent.UnmarshalEventsDynamoDBAttributeValues(image, &record)

	request.PushLogSession(
		ss.
			NewLogPrefix(
				func() []ss.LogMsgAttr {
					return ss.NewLogMsgAttrRequestDumps(request)
				}).
			Add(record.User).
			AddVal("dbevent", eventName))

	return record.User
}

func (lambda lambda) run(
	user ss.UserID,
	request dbeventlambda.Request,
	event events.DynamoDBEventRecord,
	factories ...UpdaterFactory,
) {
	var barrier sync.WaitGroup
	for _, factory := range factories {
		factory(user, request, lambda.db, &barrier).SpawnUpdate()
	}
	barrier.Wait()
}
