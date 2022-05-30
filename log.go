// Copyright 2021-2022, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ss

import (
	"log"
	"sync"
	"sync/atomic"
)

type LogSource interface{ Log() LogStream }

// Log describes product log interface.
type LogStream interface {
	NoCopy

	Debug(*LogMsg)
	Info(*LogMsg)
	Warn(*LogMsg)
	Error(*LogMsg)
	Panic(*LogMsg)

	CheckPanic(panicValue interface{}, errorMessage string)
	checkPanic(panicValue interface{}, getPanicDetails func() *LogMsg)
}

type Log interface {
	LogStream

	Started()

	// NewLogSession creates the log session which allows setting records
	// prefix for each session message.
	NewSession(newPrefix func() LogPrefix) LogSession

	// CheckExit makes final check for panic, writes all error data.
	// It has to be the one and the lowest level check in the call.
	// Each goroutine should has at the beggining sonthing like:
	// defer func() { service.log.CheckExit(recover()) }()
	CheckExit(panicValue interface{})
}

type LogSession interface {
	LogStream

	// NewLogSession creates the log session which allows setting records
	// prefix for each session message.
	NewSession(newPrefix func() LogPrefix) LogSession
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

func NewLog(projectPackage string, module string, config Config) Log {

	result := serviceLog{
		destinations: []logDestination{},
		statics: map[string]interface{}{
			"module":  module,
			"package": projectPackage,
			"build": map[string]interface{}{
				"id":      config.SS.Build.ID,
				"commit":  config.SS.Build.Commit,
				"builder": config.SS.Build.Builder,
				"env":     config.SS.Build.GetEnvironment(),
			},
			"aws": map[string]interface{}{
				"region": config.SS.Service.AWS.Region,
			},
		},
		messageChan: make(chan serviceLogMessage, 100),
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

	log.SetFlags(log.LstdFlags | log.Lmicroseconds)

	go result.runWriter()

	return &result
}

type serviceLog struct {
	NoCopyImpl

	destinations []logDestination
	sentry       sentry

	statics        map[string]interface{}
	sequenceNumber uint32

	messageChan chan serviceLogMessage
}

type serviceLogMessage struct {
	Write    func()
	SyncChan chan<- struct{}
}

func (l *serviceLog) Started() {
	if !S.Config().IsExtraLogEnabled() {
		return
	}
	l.Debug(NewLogMsg("started"))
}

func (l *serviceLog) NewSession(newPrefix func() LogPrefix) LogSession {
	return newLogSession(l, newPrefix)
}

func (l *serviceLog) CheckPanic(panicValue interface{}, errorMessage string) {
	l.checkPanic(panicValue, func() *LogMsg { return NewLogMsg(errorMessage) })
}

func (l *serviceLog) CheckExit(panicValue interface{}) {

	// This sync allows sending all messages before exiting from lambda,
	// even if it's not an error or panic - CheckExit is the last call
	// from lambda to log, and it has to wait until records will be sent.
	defer l.sync()

	message := l.checkPanicValue(
		panicValue,
		func() *LogMsg { return NewLogMsg("panic detected at exit") })
	if message == nil {
		return
	}

	l.sentry.Recover(message)
	l.forEachDestination(func(d logDestination) error {
		return d.WritePanic(message)
	})

	l.sync() //  it could be the last chance to show log messages in the queue

	log.Panicf(
		"%s %s",
		message.GetMessage(),
		message.ConvertAttributesToJSON())
}

func (l *serviceLog) checkPanicValue(
	panicValue interface{},
	getPanicDetails func() *LogMsg,
) *LogMsg {
	if panicValue == nil {
		return nil
	}

	if result, isMessage := panicValue.(*LogMsg); isMessage {
		// Collecting info from try-catch levels.
		result.AddParent(getPanicDetails())
		return result
	}

	// It's unhandled panic, and panic raised not thought log.Panic.
	result := getPanicDetails()
	result.AddPanic(panicValue)
	return result
}

func (l *serviceLog) checkPanic(
	panicValue interface{},
	getPanicDetails func() *LogMsg,
) {
	message := l.checkPanicValue(panicValue, getPanicDetails)
	if message == nil {
		return
	}
	l.panic(message)
}

func (l *serviceLog) Debug(m *LogMsg) {
	sequenceNumber := atomic.AddUint32(&l.sequenceNumber, 1)
	l.messageChan <- serviceLogMessage{
		Write: func() {
			l.setStatics(logLevelDebug, sequenceNumber, m)
			l.print(m)
			l.forEachDestination(func(d logDestination) error {
				return d.WriteDebug(m)
			})
		},
	}
}

func (l *serviceLog) Info(m *LogMsg) {
	sequenceNumber := atomic.AddUint32(&l.sequenceNumber, 1)
	l.messageChan <- serviceLogMessage{
		Write: func() {
			l.setStatics(logLevelInfo, sequenceNumber, m)
			l.print(m)
			l.forEachDestination(func(d logDestination) error {
				return d.WriteInfo(m)
			})
		},
	}
}

func (l *serviceLog) Warn(m *LogMsg) {
	l.setStatics(
		logLevelWarn,
		atomic.AddUint32(&l.sequenceNumber, 1),
		m)

	l.sentry.CaptureMessage(m)

	l.messageChan <- serviceLogMessage{
		Write: func() {
			l.print(m)
			l.forEachDestination(func(d logDestination) error {
				return d.WriteWarn(m)
			})
		},
	}
}

func (l *serviceLog) Error(m *LogMsg) {
	l.setStatics(
		logLevelError,
		atomic.AddUint32(&l.sequenceNumber, 1),
		m)
	m.AddCurrentStack()

	l.sentry.CaptureMessage(m)

	l.messageChan <- serviceLogMessage{
		Write: func() {
			l.print(m)
			l.forEachDestination(func(d logDestination) error {
				return d.WriteError(m)
			})
		},
	}
}

func (l *serviceLog) Panic(message *LogMsg) {
	l.setStatics(
		logLevelPanic,
		atomic.AddUint32(&l.sequenceNumber, 1),
		message)
	message.AddCurrentStack()

	l.panic(message)
}

func (l *serviceLog) setStatics(
	level logLevel,
	sequenceNumber uint32,
	message *LogMsg,
) {
	message.AddVal("n", sequenceNumber)
	for k, v := range l.statics {
		message.AddVal(k, v)
	}
	message.SetLevel(level)
}

func (l *serviceLog) panic(message *LogMsg) {
	// Just start proccess of panicing, to collect debug info on all levels
	// of catching.
	panic(message)
}

func (l *serviceLog) print(message *LogMsg) {
	log.Printf(
		"%s: %s %s",
		message.GetLevel(),
		message.GetMessage(),
		message.ConvertAttributesToJSON())
}

func (l *serviceLog) sync() {
	syncChan := make(chan struct{})
	l.messageChan <- serviceLogMessage{SyncChan: syncChan}
	l.sentry.Flush()
	<-syncChan
}

func (l *serviceLog) syncDestinations() {

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

	select {
	case <-doneSignalChan:
		break
	case <-S.SubscribeForLambdaTimeout():
		l.Error(NewLogMsg("log sync timeout"))
	}
}

func (l *serviceLog) forEachDestination(callback func(logDestination) error) {
	for _, d := range l.destinations {
		if err := callback(d); err != nil {
			log.Printf(
				"Failed to call log destination %q method: %v",
				d.GetName(),
				err)
		}
	}
}

func (l *serviceLog) runWriter() {
	for {
		message, isOpen := <-l.messageChan
		if !isOpen {
			break
		}
		if message.Write != nil {
			message.Write()
		}
		if message.SyncChan != nil {
			l.syncDestinations()
			message.SyncChan <- struct{}{}
		}
	}
}

////////////////////////////////////////////////////////////////////////////////

type serviceLogSession struct {
	NoCopyImpl

	log            LogStream
	newPrefix      func() LogPrefix
	prefixCache    LogPrefix
	isPrefixCached bool
}

func newLogSession(log LogStream, newPrefix func() LogPrefix) LogSession {
	return &serviceLogSession{log: log, newPrefix: newPrefix}
}

func (s *serviceLogSession) NewSession(newPrefix func() LogPrefix) LogSession {
	return newLogSession(s, newPrefix)
}

func (s *serviceLogSession) CheckPanic(
	panicValue interface{},
	errorMessage string,
) {
	s.checkPanic(panicValue, func() *LogMsg { return NewLogMsg(errorMessage) })
}

func (s *serviceLogSession) checkPanic(
	panicValue interface{},
	getPanicDetails func() *LogMsg,
) {
	s.log.checkPanic(
		panicValue,
		func() *LogMsg { return getPanicDetails().AddFailPrefix(s.getPrefix()) })
}

func (s *serviceLogSession) Debug(m *LogMsg) {
	s.log.Debug(m.AddInfoPrefix(s.getPrefix()))
}
func (s *serviceLogSession) Info(m *LogMsg) {
	s.log.Info(m.AddInfoPrefix(s.getPrefix()))
}
func (s *serviceLogSession) Warn(m *LogMsg) {
	s.log.Warn(m.AddFailPrefix(s.getPrefix()))
}
func (s *serviceLogSession) Error(m *LogMsg) {
	s.log.Error(m.AddFailPrefix(s.getPrefix()))
}
func (s *serviceLogSession) Panic(m *LogMsg) {
	s.log.Panic(m.AddFailPrefix(s.getPrefix()))
}

func (s *serviceLogSession) getPrefix() LogPrefix {
	if !s.isPrefixCached {
		s.prefixCache = s.newPrefix()
	}
	return s.prefixCache
}

////////////////////////////////////////////////////////////////////////////////
