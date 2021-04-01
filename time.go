// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ss

import (
	"time"
)

// Now return current platform time in correct time zone.
func Now() time.Time { return time.Now().UTC() }
