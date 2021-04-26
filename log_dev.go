// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

// +build windows darwin

package ss

import (
	"fmt"
	"log"
)

type serviceLog struct{ isPanic bool }

func newServiceLog(projectPackage, module string, config Config) ServiceLog {
	return &serviceLog{}
}

func (serviceLog *serviceLog) NewSession(prefix string) ServiceLogStream {
	return newLogSession(serviceLog, prefix)
}

func (serviceLog *serviceLog) CheckExit(panicValue interface{}) {
	serviceLog.checkPanic(panicValue)
}

func (serviceLog *serviceLog) CheckExitWithPanicDetails(
	panicValue interface{},
	getPanicDetails func() string,
) {
	serviceLog.checkPanicWithDetails(panicValue, getPanicDetails)
}

func (serviceLog *serviceLog) checkPanic(panicValue interface{}) {
	serviceLog.checkPanicWithDetails(panicValue, nil)
}

func (serviceLog *serviceLog) checkPanicWithDetails(
	panicValue interface{},
	getPanicDetails func() string,
) {
	if panicValue == nil {
		return
	}

	if serviceLog.isPanic {
		panic(panicValue)
	}

	if getPanicDetails == nil {
		serviceLog.Panic(`Panic detected: "%v".`, panicValue)
		return
	}
	serviceLog.Panic(
		`Panic detected: "%v". Details: %s.`,
		panicValue,
		getPanicDetails())
}

func (serviceLog serviceLog) Started() {
	serviceLog.Debug("Started %s on Windows.", S.Build().ID)
}

func (serviceLog serviceLog) Debug(format string, args ...interface{}) {
	log.Printf("Debug: "+format+"\n", args...)
}

func (serviceLog serviceLog) Info(format string, args ...interface{}) {
	log.Printf(format+"\n", args...)
}

func (serviceLog serviceLog) Warn(format string, args ...interface{}) {
	log.Printf("Warn: "+format+"\n", args...)
}

func (serviceLog serviceLog) Error(format string, args ...interface{}) {
	log.Printf("Error: "+format+"\n", args...)
}

func (serviceLog serviceLog) Err(err error) {
	serviceLog.Error(convertLogErrToString(err))
}

func (serviceLog *serviceLog) Panic(format string, args ...interface{}) {
	serviceLog.isPanic = true
	log.Panicln(fmt.Sprintf(format, args...))
}
