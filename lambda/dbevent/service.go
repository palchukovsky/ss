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
	if len(event.Records) == 0 {
		ss.S.Log().Warn("Empty event list.")
		return nil
	}

	eventID := event.Records[0].EventID
	log := ss.S.Log().NewSession(
		eventID[:4] + "-" + eventID[len(eventID)-4:])
	defer log.CheckExit(recover())

	log.Debug("%d records.", len(event.Records))
	if ss.S.Config().IsExtraLogEnabled() {
		log.Debug("Records dump: %s", ss.Dump(event.Records))
	}

	request := newRequest(event.Records, log)

	if err := service.Lambda.Execute(request); err != nil {
		request.Log().Error(`Lambda execution error: "%v". Events dump: %s`,
			err, ss.Dump(event.Records))
		return err
	}

	return nil
}
