// Copyright 2021-2022, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package dbeventlambda

// Lambda describes lambda to handle DynamoDB event.
type Lambda interface {
	Execute(Request) error
}

////////////////////////////////////////////////////////////////////////////////
