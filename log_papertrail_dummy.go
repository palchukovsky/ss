// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

// +build windows

package ss

import (
	"fmt"
	"log"
	"log/syslog"
	"strconv"
)

func newPapertrailLogIfSet(
	projectPackage,
	module string,
	config Config,
	sentry sentry,
) (logDestination, error) {
	// Doesn't have implementation for specified in "+build" OS.
	return nil
}
