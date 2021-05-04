// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package dbeventlambda

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/palchukovsky/ss"
)

// Request describes request to lambda which handles DynamoDB event.
type Request interface {
	Log() ss.LogStream
	StartLogSession(ss.LogPrefix)

	GetEvents() []events.DynamoDBEventRecord
}

////////////////////////////////////////////////////////////////////////////////

type request struct {
	log    ss.LogStream
	events []events.DynamoDBEventRecord
}

func newRequest(
	events []events.DynamoDBEventRecord,
	log ss.LogStream,
) Request {
	return &request{
		log:    log,
		events: events,
	}
}

func (request request) Log() ss.LogStream { return request.log }

func (request *request) StartLogSession(prefix ss.LogPrefix) {
	request.log = request.log.NewSession(prefix)
}

func (request request) GetEvents() []events.DynamoDBEventRecord {
	return request.events
}
