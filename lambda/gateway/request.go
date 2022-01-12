// Copyright 2021-2022, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package gatewaylambda

import (
	"encoding/json"
	"fmt"

	"github.com/palchukovsky/ss"
	"github.com/palchukovsky/ss/lambda"
)

// NewRequest creates new Request instance.
func NewRequest(
	gateway lambda.Gateway,
	log ss.LogSession,
	defaultResponse interface{},
) Request {
	return Request{
		gateway:      gateway,
		log:          log,
		ResponseBody: defaultResponse,
	}
}

// Request implements request method for output gateway.
type Request struct {
	ss.NoCopyImpl

	log          ss.LogSession
	gateway      lambda.Gateway
	ResponseBody interface{}
}

// Log returns request log session.
func (request *Request) Log() ss.LogStream { return request.log }

// Response responses to request with given data.
func (request *Request) Respond(response interface{}) {
	request.ResponseBody = response
}

// UnmarshalRequest parses request from a string.
func (request *Request) UnmarshalRequest(
	source string,
	result interface{},
) error {
	if err := json.Unmarshal([]byte(source), result); err != nil {
		return fmt.Errorf(`failed to parse request: "%w"`, err)
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
