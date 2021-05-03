// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ss

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"regexp"
	"runtime"
	"strings"
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

type logDestination interface {
	GetName() string

	WriteDebug(message string) error
	WriteInfo(message string) error
	WriteWarn(message string) error
	WriteError(message string) error
	WritePanic(message string) error

	Sync() error
}

////////////////////////////////////////////////////////////////////////////////

func newServiceLog(
	projectPackage string,
	module string,
	config Config,
) ServiceLog {

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
	build := S.Build()
	l.Debug(
		"Started %q ver %q on %q",
		build.ID,
		build.Version,
		S.Config().AWS.Region)
}

func (l *serviceLog) NewSession(prefix string) ServiceLogStream {
	return newLogSession(l, prefix)
}

func (l *serviceLog) CheckExit(panicValue interface{}) {
	l.CheckExitWithPanicDetails(panicValue, nil)
}

func (l *serviceLog) CheckExitWithPanicDetails(
	panicValue interface{},
	getPanicDetails func() string,
) {
	defer l.sentry.Flush()
	defer l.sync()
	l.checkPanicWithDetails(panicValue, getPanicDetails)
}

func (l *serviceLog) checkPanic(panicValue interface{}) {
	l.checkPanicWithDetails(panicValue, nil)
}

func (l *serviceLog) checkPanicWithDetails(
	panicValue interface{},
	getPanicDetails func() string,
) {
	if panicValue == nil {
		return
	}

	if l.isPanic {
		// Rethrow.
		panic(panicValue)
	}

	message := fmt.Sprintf(`Panic detected: %v`, panicValue)
	if getPanicDetails != nil {
		message += fmt.Sprintf(`. Details: %s`, getPanicDetails())
	}

	l.panic(panicValue, message)
}

func (l serviceLog) Debug(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	l.forEachDestination(func(d logDestination) error {
		return d.WriteDebug(message)
	})
}

func (l serviceLog) Info(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	log.Println("Info: " + message)
	l.forEachDestination(func(d logDestination) error {
		return d.WriteInfo(message)
	})
}

func (l serviceLog) Warn(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)

	log.Println("Warn: " + message)

	if err := l.sentry.CaptureMessage(l.removePrefix(message)); err != nil {
		l.Error("Failed capture message by Sentry: %v", err)
	}

	l.forEachDestination(func(d logDestination) error {
		return d.WriteWarn(message)
	})
}

func (l serviceLog) Error(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)

	log.Println("Error: " + message)

	l.sentry.CaptureException(errors.New(l.removePrefix(message)))

	l.forEachDestination(func(d logDestination) error {
		return d.WriteError(message)
	})
}

func (l serviceLog) Err(err error) {
	message := capitalizeString(fmt.Sprintf("%v", err))

	log.Println("Error: " + message)

	l.sentry.CaptureException(err)

	l.forEachDestination(func(d logDestination) error {
		return d.WriteError(message)
	})
}

func (l serviceLog) Panic(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	l.panic(message, message)
}

func (l *serviceLog) panic(panicValue interface{}, message string) {
	defer l.sentry.Flush()
	defer l.sync()

	l.sentry.Recover(panicValue)

	l.writePanic(message)

	l.isPanic = true
	log.Panicln(message)

}

func (l *serviceLog) writePanic(message string) {
	message = "Panic! " + message + ". Stack: ["

	{
		buffer := make([]byte, 4096)
		size := runtime.Stack(buffer, false)
		isStared := false
		scanner := bufio.NewScanner(strings.NewReader(string(buffer[:size])))
		for scanner.Scan() {
			if isStared {
				message += ","
			} else {
				isStared = true
			}
			if scanner.Text()[:1] == "\t" {
				message += fmt.Sprintf("%q", "    "+scanner.Text()[1:])
			} else {
				message += fmt.Sprintf("%q", scanner.Text())
			}
		}
		message += "]"
	}

	l.forEachDestination(func(d logDestination) error {
		return d.WritePanic(message)
	})
}

func (l serviceLog) sync() {
	l.forEachDestination(func(l logDestination) error {
		return l.Sync()
	})
}

func (l serviceLog) forEachDestination(callback func(logDestination) error) {
	defer func() {
		if err := recover(); err != nil {
			log.Panicf("Failed to call log destination method: %v", err)
		}
	}()
	for _, d := range l.destinations {
		if err := callback(d); err != nil {
			log.Printf(
				"Failed to call log destination %q method: %v",
				d.GetName(),
				err)
		}
	}
}

func (serviceLog) removePrefix(source string) string {
	return string(
		regexp.MustCompile(
			`(?m)^(\[[^\]]*\]\s*)*(.+)$`).
			ReplaceAll(
				[]byte(source),
				[]byte("$2")))
}

////////////////////////////////////////////////////////////////////////////////

type serviceLogSession struct {
	log    ServiceLogStream
	prefix string
}

func newLogSession(log ServiceLogStream, prefix string) ServiceLogStream {
	return &serviceLogSession{log: log, prefix: "[" + prefix + "] "}
}

func (log *serviceLogSession) NewSession(prefix string) ServiceLogStream {
	return newLogSession(log, prefix)
}

func (log *serviceLogSession) CheckExit(panicValue interface{}) {
	log.checkPanic(panicValue)
}

func (log *serviceLogSession) CheckExitWithPanicDetails(
	panicValue interface{},
	getPanicDetails func() string,
) {
	log.checkPanicWithDetails(panicValue, getPanicDetails)
}

func (log *serviceLogSession) checkPanic(panicValue interface{}) {
	log.log.checkPanic(panicValue)
}

func (log *serviceLogSession) checkPanicWithDetails(
	panicValue interface{},
	getPanicDetails func() string,
) {
	log.log.checkPanicWithDetails(panicValue, getPanicDetails)
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
	log.Error(capitalizeString(fmt.Sprintf("%v", err)))
}

func (log *serviceLogSession) Panic(format string, args ...interface{}) {
	log.log.Panic(log.prefix+format, args...)
}

////////////////////////////////////////////////////////////////////////////////
