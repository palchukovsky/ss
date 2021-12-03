// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package dbeventlambda

import (
	"github.com/aws/aws-lambda-go/events"
	awslambda "github.com/aws/aws-lambda-go/lambda"
	"github.com/palchukovsky/ss"
	"github.com/palchukovsky/ss/lambda"
)

// NewService creates new lambda service instance to work with DynamoDB stream.
func NewService(lambda Lambda) lambda.Service {
	result := &service{Lambda: lambda}
	ss.S.Log().Started()
	return result
}

type service struct{ Lambda }

func (service *service) Start() {
	awslambda.Start(
		func(event *events.DynamoDBEvent) error {
			service.handle(event)
			return nil
		})
}

func (service *service) handle(event *events.DynamoDBEvent) {
	ss.S.StartLambda(
		func() []ss.LogMsgAttr {
			return ss.NewLogMsgAttrRequestDumps(*event)
		})
	defer func() { ss.S.CompleteLambda(recover()) }()

	log := ss.S.Log().NewSession(
		ss.
			NewLogPrefix(
				func() []ss.LogMsgAttr {
					return ss.NewLogMsgAttrRequestDumps(*event)
				}).
			AddRequestID(event.Records[0].EventID))
	defer func() { log.CheckPanic(recover(), "panic at DB-event handling") }()

	if len(event.Records) == 0 {
		ss.S.Log().Panic(ss.NewLogMsg("empty event list"))
	}

	if ss.S.Config().IsExtraLogEnabled() {
		ss.S.Log().Debug(
			ss.
				NewLogMsg("event with %d records", len(event.Records)).
				AddRequest(*event))
	}

	request := newRequest(event.Records, log)

	if err := service.Lambda.Execute(request); err != nil {
		request.Log().Panic(
			ss.
				NewLogMsg(`lambda execution error`).
				AddErr(err))
	}
}
