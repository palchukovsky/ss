// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ss

import (
	"fmt"
)

// ServiceLog describes product log interface.
type ServiceLog interface {
	// NewLogSession creates the log session which allows setting records
	// prefix for each session message.
	NewSession(prefix string) ServiceLog

	CheckExit(panicErr interface{}, getPanicDetails func() string)

	Started()

	Debug(format string, args ...interface{})

	Info(format string, args ...interface{})

	Warn(format string, args ...interface{})

	Error(format string, args ...interface{})
	Err(err error)

	Panic(format string, args ...interface{})
}

////////////////////////////////////////////////////////////////////////////////

type serviceLogSession struct {
	log    ServiceLog
	prefix string
}

func newLogSession(log ServiceLog, prefix string) ServiceLog {
	return &serviceLogSession{log: log, prefix: "[" + prefix + "] "}
}

func (log *serviceLogSession) NewSession(prefix string) ServiceLog {
	return newLogSession(log, prefix)
}

func (log *serviceLogSession) CheckExit(
	panicErr interface{},
	getPanicDetails func() string,
) {
	log.log.CheckExit(panicErr, getPanicDetails)
}

func (log serviceLogSession) Started() { log.log.Started() }

func (log serviceLogSession) Debug(format string, args ...interface{}) {
	log.log.Debug(log.prefix+format, args...)
}

func (log serviceLogSession) Info(format string, args ...interface{}) {
	log.log.Info(log.prefix+format, args...)
}

func (log serviceLogSession) Warn(format string, args ...interface{}) {
	log.log.Warn(log.prefix+format, args...)
}

func (log serviceLogSession) Error(format string, args ...interface{}) {
	log.log.Error(log.prefix+format, args...)
}

func (log serviceLogSession) Err(err error) {
	log.Error(convertLogErrToString(err))
}

func (log *serviceLogSession) Panic(format string, args ...interface{}) {
	log.log.Panic(log.prefix+format, args...)
}

////////////////////////////////////////////////////////////////////////////////

func convertLogErrToString(err error) string {
	result := capitalizeString(fmt.Sprintf("%v", err))
	if len(result) > 0 && result[len(result)-1:] != "." {
		result += "."
	}
	return result
}

////////////////////////////////////////////////////////////////////////////////
