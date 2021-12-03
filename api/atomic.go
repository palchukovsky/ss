// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package api

import (
	"fmt"

	"github.com/palchukovsky/ss"
)

// RunAtomic runs operation which has to be atomic.
func RunAtomic(operation func() (isCompleted bool), log ss.LogSource) error {
	count := 0
	for ; count < 10; count++ {
		if isCompleted := operation(); isCompleted {
			return nil
		}
		log.Log().Debug(
			ss.NewLogMsg("repeating atomic operation %d time...", count+2))
	}
	return fmt.Errorf(`failed to execute atomic operation %d times`, count+1)
}
