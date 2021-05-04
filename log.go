// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ss

import (
	"log"
	"sync"
	"time"
)

// Log describes product log interface.
type LogStream interface {
	// NewLogSession creates the log session which allows setting records
	// prefix for each session message.
	NewSession(LogPrefix) LogStream

	CheckExit(panicValue interface{})
	CheckExitWithPanicDetails(
		panicValue interface{},
		getPanicDetails func() *LogMsg)

	checkPanic(panicValue interface{})
	checkPanicWithDetails(
		panicValue interface{},
		getPanicDetails func() *LogMsg)

	Debug(*LogMsg)
	Info(*LogMsg)
	Warn(*LogMsg)
	Error(*LogMsg)
	Panic(*LogMsg)
}

type Log interface {
	LogStream

	Started()
}

////////////////////////////////////////////////////////////////////////////////

type logLevel string

const (
	logLevelDebug logLevel = "debug"
	logLevelInfo  logLevel = "info"
	logLevelWarn  logLevel = "warn"
	logLevelError logLevel = "error"
	logLevelPanic logLevel = "panic"
)

////////////////////////////////////////////////////////////////////////////////

type logDestination interface {
	GetName() string

	WriteDebug(*LogMsg) error
	WriteInfo(*LogMsg) error
	WriteWarn(*LogMsg) error
	WriteError(*LogMsg) error
	WritePanic(*LogMsg) error

	Sync() error
}

////////////////////////////////////////////////////////////////////////////////

func NewLog(
	projectPackage string,
	module string,
	config Config,
) Log {

	result := serviceLog{
		destinations: []logDestination{},
	}

	{
		var err error
		result.sentry, err = newSentry(projectPackage, module, config)
		if err != nil {
			log.Panicf("Failed to init Sentry: %v", err)
		}
	}

	{
		loggly, err := newLogglyIfSet(
			projectPackage,
			module,
			config,
			result.sentry)
		if err != nil {
			log.Panicf("Failed to init Loggly: %v", err)
		}
		if loggly != nil {
			result.destinations = append(result.destinations, loggly)
		}
	}

	{
		logzio, err := newLogzioIfSet(
			projectPackage,
			module,
			config,
			result.sentry)
		if err != nil {
			log.Panicf("Failed to init Logz.io: %v", err)
		}
		if logzio != nil {
			result.destinations = append(result.destinations, logzio)
		}
	}

	{
		papertrail, err := newPapertrailIfSet(
			projectPackage,
			module,
			config,
			result.sentry)
		if err != nil {
			log.Panicf("Failed to init Papertail: %v", err)
		}
		if papertrail != nil {
			result.destinations = append(result.destinations, papertrail)
		}
	}

	return &result
}

type serviceLog struct {
	destinations []logDestination
	sentry       sentry
	isPanic      bool
}

func (l serviceLog) Started() {
	l.Debug(NewLogMsg("started"))
}

func (l *serviceLog) NewSession(prefix LogPrefix) LogStream {
	return newLogSession(l, prefix)
}

func (l *serviceLog) CheckExit(panicValue interface{}) {
	l.CheckExitWithPanicDetails(
		panicValue,
		func() *LogMsg { return NewLogMsg("panic detected at exit") })
}

func (l *serviceLog) CheckExitWithPanicDetails(
	panicValue interface{},
	getPanicDetails func() *LogMsg,
) {
	defer l.sentry.Flush()
	defer l.sync()
	l.checkPanicWithDetails(panicValue, getPanicDetails)
}

func (l *serviceLog) checkPanic(panicValue interface{}) {
	l.checkPanicWithDetails(
		panicValue,
		func() *LogMsg { return NewLogMsg("panic detected") })
}

func (l *serviceLog) checkPanicWithDetails(
	panicValue interface{},
	getPanicDetails func() *LogMsg,
) {
	if panicValue == nil {
		return
	}
	if l.isPanic {
		// Rethrow.
		panic(panicValue)
	}
	l.panic(panicValue, getPanicDetails())
}

func (l serviceLog) Debug(m *LogMsg) {
	m.SetLevel(logLevelDebug)
	l.print(m)
	l.forEachDestination(func(d logDestination) error { return d.WriteDebug(m) })
}

func (l serviceLog) Info(m *LogMsg) {
	m.SetLevel(logLevelInfo)
	l.print(m)
	l.forEachDestination(func(d logDestination) error { return d.WriteInfo(m) })
}

func (l serviceLog) Warn(m *LogMsg) {
	m.SetLevel(logLevelWarn)
	l.print(m)
	l.sentry.CaptureMessage(m)
	l.forEachDestination(func(d logDestination) error { return d.WriteWarn(m) })
}

func (l serviceLog) Error(m *LogMsg) {
	m.SetLevel(logLevelError)
	m.AddCurrentStack()
	l.print(m)
	l.sentry.CaptureMessage(m)
	l.forEachDestination(func(d logDestination) error { return d.WriteError(m) })
}

func (l serviceLog) Panic(m *LogMsg) {
	m.SetLevel(logLevelPanic)
	l.panic(m.GetMessage(), m)
}

func (l *serviceLog) panic(panicValue interface{}, message *LogMsg) {
	defer l.sentry.Flush()
	defer l.sync()

	message.AddCurrentStack()

	l.sentry.Recover(panicValue, message)
	l.forEachDestination(func(d logDestination) error {
		return d.WritePanic(message)
	})

	l.isPanic = true
	log.Panicln(message)
}

func (l serviceLog) print(message *LogMsg) {
	log.Printf(
		"%s: %s %s",
		message.GetLevel(),
		message.GetMessage(),
		message.ConvertAttributesToJSON())
}

func (l serviceLog) sync() {

	var wait sync.WaitGroup
	l.forEachDestination(func(d logDestination) error {
		wait.Add(1)
		go func() {
			err := d.Sync()
			wait.Done()
			if err != nil {
				l.Error(NewLogMsg("failed to sync log  %q", d.GetName()).AddErr(err))
			}
		}()
		return nil
	})

	doneSignalChan := make(chan struct{})
	defer close(doneSignalChan)
	go func() {
		wait.Wait()
		doneSignalChan <- struct{}{}
	}()

	// Common timeout for all logs, as lambda has runtime time limit.
	timeoutChan := time.After(2750 * time.Millisecond)
	select {
	case <-doneSignalChan:
		break
	case <-timeoutChan:
		l.Error(NewLogMsg("log sync timeout"))
	}
}

func (l serviceLog) forEachDestination(callback func(logDestination) error) {
	for _, d := range l.destinations {
		if err := callback(d); err != nil {
			log.Printf(
				"Failed to call log destination %q method: %v",
				d.GetName(),
				err)
		}
	}
}

////////////////////////////////////////////////////////////////////////////////

type serviceLogSession struct {
	log    LogStream
	prefix LogPrefix
}

func newLogSession(
	log LogStream,
	prefix LogPrefix,
) LogStream {
	return &serviceLogSession{log: log, prefix: prefix}
}

func (s *serviceLogSession) NewSession(prefix LogPrefix) LogStream {
	return newLogSession(s, prefix)
}

func (s serviceLogSession) CheckExit(panicValue interface{}) {
	s.checkPanic(panicValue)
}

func (s serviceLogSession) CheckExitWithPanicDetails(
	panicValue interface{},
	getPanicDetails func() *LogMsg,
) {
	s.checkPanicWithDetails(
		panicValue,
		func() *LogMsg { return getPanicDetails().AddPrefix(s.prefix) })
}

func (s serviceLogSession) checkPanic(panicValue interface{}) {
	s.log.checkPanic(panicValue)
}

func (s serviceLogSession) checkPanicWithDetails(
	panicValue interface{},
	getPanicDetails func() *LogMsg,
) {
	s.log.checkPanicWithDetails(
		panicValue,
		func() *LogMsg { return getPanicDetails().AddPrefix(s.prefix) })
}

func (s serviceLogSession) Debug(m *LogMsg) {
	s.log.Debug(m.AddPrefix(s.prefix))
}
func (s serviceLogSession) Info(m *LogMsg) {
	s.log.Info(m.AddPrefix(s.prefix))
}
func (s serviceLogSession) Warn(m *LogMsg) {
	s.log.Warn(m.AddPrefix(s.prefix))
}
func (s serviceLogSession) Error(m *LogMsg) {
	s.log.Error(m.AddPrefix(s.prefix))
}
func (s serviceLogSession) Panic(m *LogMsg) {
	s.log.Panic(m.AddPrefix(s.prefix))
}

////////////////////////////////////////////////////////////////////////////////
