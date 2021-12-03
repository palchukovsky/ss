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
	PushLogSession(ss.LogPrefix)
	// PopLogSession pops log session. It has to be called in the stack end,
	// only if no error occurred. If it is called with "defer" - an unhandled
	// panic will not have all session data to store info about the error.
	PopLogSession(panicValue interface{})

	GetEvents() []events.DynamoDBEventRecord
}

////////////////////////////////////////////////////////////////////////////////

type request struct {
	log    []ss.LogSession
	events []events.DynamoDBEventRecord
}

func newRequest(
	events []events.DynamoDBEventRecord,
	log ss.LogSession,
) Request {
	return &request{
		log:    []ss.LogSession{log},
		events: events,
	}
}

func (request request) Log() ss.LogStream {
	return request.log[len(request.log)-1]
}

func (request *request) PushLogSession(prefix ss.LogPrefix) {
	request.log = append(
		request.log,
		request.log[len(request.log)-1].NewSession(prefix))
}

func (request *request) PopLogSession(panicValue interface{}) {
	// If panic - it uses last pushed log session to collect debug info
	request.Log().CheckPanic(panicValue, "specific request handling panic")
	// No panic - last log session is not required.
	request.log = request.log[:len(request.log)-1]
}

func (request request) GetEvents() []events.DynamoDBEventRecord {
	return request.events
}
