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

func (service *service) Start() { awslambda.Start(service.handle) }

func (service *service) handle(event *events.DynamoDBEvent) error {
	defer func() { ss.S.Log().CheckExit(recover()) }()

	if len(event.Records) == 0 {
		ss.S.Log().Warn(ss.NewLogMsg("empty event list").AddRequest(*event))
		return nil
	}

	log := ss.S.Log().NewSession(
		ss.NewLogPrefix().AddRequestID(event.Records[0].EventID))
	defer func() {
		log.CheckPanic(
			recover(),
			func() *ss.LogMsg {
				return ss.NewLogMsg("panic at service handling").AddRequest(*event)
			})
	}()

	log.Debug(ss.NewLogMsg("%d records", len(event.Records)))

	request := newRequest(event.Records, log)

	if err := service.Lambda.Execute(request); err != nil {
		request.Log().Error(
			ss.
				NewLogMsg(`lambda execution error`).
				AddErr(err).
				AddRequest(*event))
		return err
	}

	return nil
}
