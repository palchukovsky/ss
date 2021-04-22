// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ss

import (
	"fmt"
	"log"
)

func NewServiceDevLog(projectPackage, module string) ServiceLog {
	return &serviceDevLog{}
}

type serviceDevLog struct{ isPanic bool }

func (serviceLog *serviceDevLog) NewSession(prefix string) ServiceLog {
	return newLogSession(serviceLog, prefix)
}

func (serviceLog *serviceDevLog) CheckExit() {
	if serviceLog.isPanic {
		return
	}
	if err := recover(); err != nil {
		serviceLog.Panic(`Panic detected: "%v".`, err)
	}
}

func (serviceLog serviceDevLog) Started() {
	serviceLog.Debug("Started %s on Windows.", S.Build().ID)
}

func (serviceLog serviceDevLog) Debug(format string, args ...interface{}) {
	log.Printf("Debug: "+format+"\n", args...)
}

func (serviceLog serviceDevLog) Info(format string, args ...interface{}) {
	log.Printf(format+"\n", args...)
}

func (serviceLog serviceDevLog) Warn(format string, args ...interface{}) {
	log.Printf("Warn: "+format+"\n", args...)
}

func (serviceLog serviceDevLog) Error(format string, args ...interface{}) {
	log.Printf("Error: "+format+"\n", args...)
}

func (serviceLog serviceDevLog) Err(err error) {
	serviceLog.Error(convertLogErrToString(err))
}

func (serviceLog *serviceDevLog) Panic(format string, args ...interface{}) {
	serviceLog.isPanic = true
	log.Panicln(fmt.Sprintf(format, args...))
}
