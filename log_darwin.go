// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

// +build darwin

package ss

func newServiceLog(projectPackage, module string, config Config) ServiceLog {
	return NewServiceDevLog(projectPackage, module)
}
