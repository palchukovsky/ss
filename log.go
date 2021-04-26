// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ss

import (
	"fmt"
)

// ServiceLog describes product log interface.
type ServiceLogStream interface {
	// NewLogSession creates the log session which allows setting records
	// prefix for each session message.
	NewSession(prefix string) ServiceLogStream

	CheckExit(panicValue interface{})
	CheckExitWithPanicDetails(
		panicValue interface{},
		getPanicDetails func() string)

	checkPanic(panicValue interface{})
	checkPanicWithDetails(panicValue interface{}, getPanicDetails func() string)

	Debug(format string, args ...interface{})

	Info(format string, args ...interface{})

	Warn(format string, args ...interface{})

	Error(format string, args ...interface{})
	Err(err error)

	Panic(format string, args ...interface{})
}

type ServiceLog interface {
	ServiceLogStream

	Started()
}

////////////////////////////////////////////////////////////////////////////////

type serviceLogSession struct {
	log      ServiceLogStream
	prefix   string
	isClosed bool
}

func newLogSession(log ServiceLogStream, prefix string) ServiceLogStream {
	return &serviceLogSession{log: log, prefix: "[" + prefix + "] "}
}

func (log *serviceLogSession) NewSession(prefix string) ServiceLogStream {
	return newLogSession(log, prefix)
}

func (log *serviceLogSession) CheckExit(panicValue interface{}) {
	log.close()
	log.checkPanic(panicValue)
}

func (log *serviceLogSession) CheckExitWithPanicDetails(
	panicValue interface{},
	getPanicDetails func() string,
) {
	log.close()
	log.checkPanicWithDetails(panicValue, getPanicDetails)
}

func (log *serviceLogSession) checkPanic(panicValue interface{}) {
	log.log.checkPanic(panicValue)
}
func (log *serviceLogSession) checkPanicWithDetails(
	panicValue interface{},
	getPanicDetails func() string) {
	log.log.checkPanicWithDetails(panicValue, getPanicDetails)
}

func (log *serviceLogSession) close() {
	if log.isClosed {
		log.Error("Log session is already closed.")
		return
	}
	log.isClosed = true
}

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
