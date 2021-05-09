// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ss

import (
	"log"
	"sync"
	"sync/atomic"
)

// Log describes product log interface.
type LogStream interface {
	// NewLogSession creates the log session which allows setting records
	// prefix for each session message.
	NewSession(LogPrefix) LogSession

	Debug(*LogMsg)
	Info(*LogMsg)
	Warn(*LogMsg)
	Error(*LogMsg)
	Panic(*LogMsg)

	checkPanic(
		panicValue interface{},
		getPanicDetails func() *LogMsg)
}

type Log interface {
	LogStream
	Started()
	CheckExit(panicValue interface{})
}

type LogSession interface {
	LogStream
	CheckPanic(
		panicValue interface{},
		getPanicDetails func() *LogMsg)
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
		statics: map[string]interface{}{
			"module":  module,
			"package": projectPackage,
			"build": map[string]interface{}{
				"id":         config.SS.Build.ID,
				"commit":     config.SS.Build.Commit,
				"builder":    config.SS.Build.Builder,
				"maintainer": config.SS.Build.Maintainer,
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
	NoCopy

	destinations []logDestination
	sentry       sentry

	isPanic int32

	statics        map[string]interface{}
	sequenceNumber uint32

	messageChan chan serviceLogMessage
}

type serviceLogMessage struct {
	Write    func()
	SyncChan chan<- struct{}
}

func (l *serviceLog) Started() {
	l.Debug(NewLogMsg("started"))
}

func (l *serviceLog) NewSession(prefix LogPrefix) LogSession {
	return newLogSession(l, prefix)
}

func (l *serviceLog) CheckExit(panicValue interface{}) {
	defer l.sync()
	l.checkPanic(
		panicValue,
		func() *LogMsg { return NewLogMsg("panic detected at exit") })
}

func (l *serviceLog) checkPanic(
	panicValue interface{},
	getPanicDetails func() *LogMsg,
) {
	if panicValue == nil {
		return
	}
	if atomic.LoadInt32(&l.isPanic) != 0 {
		// Rethrow.
		panic(panicValue)
	}
	l.panic(panicValue, getPanicDetails())
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

func (l *serviceLog) Panic(m *LogMsg) { l.panic(m.GetMessage(), m) }

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

func (l *serviceLog) panic(panicValue interface{}, message *LogMsg) {
	defer l.sync()

	l.setStatics(
		logLevelPanic,
		atomic.AddUint32(&l.sequenceNumber, 1),
		message)
	message.AddCurrentStack()
	message.AddPanic(panicValue)

	l.sentry.Recover(panicValue, message)
	l.forEachDestination(func(d logDestination) error {
		return d.WritePanic(message)
	})

	atomic.StoreInt32(&l.isPanic, 1)
	log.Panicln(message)
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
	case <-S.GetLambdaTimeout():
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
	log    LogStream
	prefix LogPrefix
}

func newLogSession(
	log LogStream,
	prefix LogPrefix,
) LogSession {
	return &serviceLogSession{log: log, prefix: prefix}
}

func (s *serviceLogSession) NewSession(prefix LogPrefix) LogSession {
	return newLogSession(s, prefix)
}

func (s serviceLogSession) CheckPanic(
	panicValue interface{},
	getPanicDetails func() *LogMsg,
) {
	s.checkPanic(panicValue, getPanicDetails)
}

func (s serviceLogSession) checkPanic(
	panicValue interface{},
	getPanicDetails func() *LogMsg,
) {
	s.log.checkPanic(
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
