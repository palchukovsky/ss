// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

// +build linux

package ss

import (
	"bufio"
	"fmt"
	"log"
	"log/syslog"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/getsentry/sentry-go"
)

type serviceLog struct {
	messageChan chan logMessage
	syncChan    chan struct{}
	writer      *syslog.Writer
	isPanic     bool
}

func newServiceLog(projectPackage, module string, config Config) ServiceLog {
	environment := "production"
	if !config.SS.Build.IsProd() {
		environment = "development"
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn:              config.SS.Log.Sentry,
		AttachStacktrace: true,
		Release:          config.SS.Build.Version,
		Environment:      environment,
		BeforeSend: func(event *sentry.Event, hint *sentry.EventHint,
		) *sentry.Event {
			event.Tags["module"] = module
			event.Tags["package"] = projectPackage
			event.Tags["build"] = config.SS.Build.ID
			event.Tags["commit"] = config.SS.Build.Commit
			event.Tags["builder"] = config.SS.Build.Builder
			event.Tags["maintainer"] = config.SS.Build.Maintainer
			event.Tags["aws.region"] = config.SS.Service.AWS.Region
			return event
		},
	})
	if err != nil {
		log.Panicf(`Failed to init Sentry "%v".`, err)
	}

	result := serviceLog{
		messageChan: make(chan logMessage, 3),
		syncChan:    make(chan struct{}),
		isPanic:     false,
	}

	result.writer, err = syslog.Dial(
		"udp",
		config.SS.Log.Papertrail,
		syslog.LOG_EMERG|syslog.LOG_KERN,
		fmt.Sprintf("%s/%s@%s", projectPackage, module, config.SS.Build.Version))
	if err != nil {
		sentry.CaptureException(err)
		result.flushSentry()
		log.Panicf(`Failed to dial syslog at address %q: "%v".`,
			config.SS.Log.Papertrail, err)
	}
	go result.runWriter()

	return &result
}

type logMessage struct {
	Write   func(w *syslog.Writer, m string) error
	Message string
	Type    string
}

func (log *serviceLog) NewSession(prefix string) ServiceLogStream {
	return newLogSession(log, prefix)
}

func (serviceLog *serviceLog) CheckExit(panicValue interface{}) {
	serviceLog.CheckExitWithPanicDetails(panicValue, nil)
}

func (serviceLog *serviceLog) CheckExitWithPanicDetails(
	panicValue interface{},
	getPanicDetails func() string,
) {
	defer serviceLog.flushSentry()
	serviceLog.syncLog()

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
		return
	}

	defer serviceLog.flushSentry()
	sentry.CurrentHub().Recover(panicValue)

	if panicValue == nil {
		serviceLog.Panic(`Panic detected: "%v".`, panicValue)
		return
	}
	serviceLog.Panic(
		`Panic detected: "%v". Details: %s.`,
		panicValue,
		getPanicDetails())
}

func (serviceLog serviceLog) Started() {
	build := S.Build()
	serviceLog.Debug("Started %q ver %q on %q.",
		build.ID, build.Version, S.Config().AWS.Region)
}

func (serviceLog serviceLog) Debug(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	serviceLog.messageChan <- logMessage{
		Write:   func(w *syslog.Writer, m string) error { return w.Debug(m) },
		Message: message,
		Type:    "Debug",
	}
}

func (serviceLog serviceLog) Info(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	serviceLog.messageChan <- logMessage{
		Write:   func(w *syslog.Writer, m string) error { return w.Info(m) },
		Message: message,
		Type:    "Info",
	}
}

func (serviceLog serviceLog) Warn(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	sentry.CaptureMessage(serviceLog.removePrefix(message))
	serviceLog.messageChan <- logMessage{
		Write:   func(w *syslog.Writer, m string) error { return w.Warning(m) },
		Message: message,
		Type:    "Warn",
	}
}

func (serviceLog serviceLog) Error(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)
	sentry.CaptureMessage(serviceLog.removePrefix(message))
	serviceLog.error(message)
}

func (serviceLog serviceLog) Err(err error) {
	sentry.CaptureException(err)
	serviceLog.error(convertLogErrToString(err))
}

func (serviceLog serviceLog) error(message string) {
	serviceLog.messageChan <- logMessage{
		Write:   func(w *syslog.Writer, m string) error { return w.Err(m) },
		Message: message,
		Type:    "Error",
	}
}

func (serviceLog *serviceLog) Panic(format string, args ...interface{}) {
	message := fmt.Sprintf(format, args...)

	sentry.CaptureMessage(serviceLog.removePrefix(message))

	serviceLog.syncLog()
	{
		messageWithStack := "Panic! " + message + " Stack: ["
		buffer := make([]byte, 4096)
		size := runtime.Stack(buffer, false)
		isStared := false
		scanner := bufio.NewScanner(strings.NewReader(string(buffer[:size])))
		for scanner.Scan() {
			if isStared {
				messageWithStack += ","
			} else {
				isStared = true
			}
			if scanner.Text()[:1] == "\t" {
				messageWithStack += fmt.Sprintf("%q", "    "+scanner.Text()[1:])
			} else {
				messageWithStack += fmt.Sprintf("%q", scanner.Text())
			}
		}
		messageWithStack += "]"
		err := serviceLog.writer.Emerg(messageWithStack)
		if err != nil {
			errMessage := fmt.Sprintf(`Error: Failed to write log record: "%v"`, err)
			log.Println(errMessage)
			sentry.CaptureMessage(errMessage)
		}
	}
	serviceLog.isPanic = true
	log.Panicln(message)
}

func (serviceLog serviceLog) runWriter() {
	sequenceNumber := 0
	for {
		message, isOpen := <-serviceLog.messageChan
		if !isOpen {
			return
		}
		if message.Write == nil {
			serviceLog.syncChan <- struct{}{}
			continue
		}
		messageText := message.Message + " " + strconv.Itoa(sequenceNumber)
		sequenceNumber++
		log.Println(message.Type + ": " + messageText)
		err := message.Write(serviceLog.writer, messageText)
		if err != nil {
			errMessage := fmt.Sprintf(`Error: Failed to write log record: "%v"\n`,
				err)
			log.Println(errMessage)
			sentry.CaptureMessage(errMessage)
		}
	}
}

func (serviceLog) removePrefix(source string) string {
	return string(regexp.MustCompile(`(?m)^(\[[^\]]*\]\s*)*(.+)$`).
		ReplaceAll([]byte(source), []byte("$2")))
}

func (serviceLog *serviceLog) flushSentry() {
	if !sentry.Flush(2 * time.Second) {
		log.Println("Not all Sentry records were flushed, timeout was reached.")
	}
}

func (serviceLog serviceLog) syncLog() {
	serviceLog.messageChan <- logMessage{}
	<-serviceLog.syncChan
}
