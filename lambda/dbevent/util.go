// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package dbeventlambda

import (
	"fmt"

	"github.com/palchukovsky/ss"
)

func NewDBEventError(sourceErr error, eventIndex int, request Request) error {
	return fmt.Errorf(`failed to handle event %d of %d: "%w", event: %s`,
		eventIndex+1,
		len(request.GetEvents()),
		sourceErr,
		ss.Dump(request.GetEvents()[eventIndex]))
}
