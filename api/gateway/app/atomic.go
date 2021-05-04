// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package apiapp

import (
	"fmt"

	"github.com/palchukovsky/ss"
	ws "github.com/palchukovsky/ss/lambda/gateway/ws"
)

// RunAtomic runs operation which has to be atomic.
func RunAtomic(operation func() (bool, error), request ws.Request) error {
	count := 0
	for ; count < 10; count++ {
		if isCompleted, err := operation(); err != nil || isCompleted {
			return err
		}
		request.Log().Debug(
			ss.NewLogMsg("repeating atomic operation %d time...", count+2))
	}
	return fmt.Errorf(`failed to execute atomic operation %d times`, count+1)
}
