// Copyright 2021-2022, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

//go:build windows
// +build windows

package ss

func newPapertrailIfSet(
	projectPackage,
	module string,
	config Config,
	sentry sentry,
) (logDestination, error) {
	// Doesn't have implementation for specified in "+build" OS.
	return nil, nil
}
