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

func (serviceLog *serviceLog) NewSession(prefix string) ServiceLog {
	return newLogSession(serviceLog, prefix)
}

func (serviceLog *serviceLog) CheckExit() {
	if serviceLog.isPanic {
		return
	}
	if err := recover(); err != nil {
		serviceLog.Panic(`Panic detected: "%v".`, err)
	}
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
