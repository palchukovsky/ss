// Copyright 2021-2022, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

//go:build darwin || linux
// +build darwin linux

package ss

import (
	"fmt"
	"log"
	"log/syslog"
	"strconv"
)

func newPapertrailIfSet(
	projectPackage,
	module string,
	config Config,
	sentry sentry,
) (logDestination, error) {

	if config.SS.Log.Papertrail == "" {
		return nil, nil
	}

	result := papertrail{
		messageChan: make(chan papertrailMessage, 10),
		syncChan:    make(chan struct{}),
		sentry:      sentry,
	}

	var err error
	result.writer, err = syslog.Dial(
		"udp",
		config.SS.Log.Papertrail,
		syslog.LOG_EMERG|syslog.LOG_KERN,
		fmt.Sprintf("%s/%s@%s", projectPackage, module, config.SS.Build.Version))
	if err != nil {
		return nil, err
	}

	go result.runWriter()

	return result, nil
}

type papertrail struct {
	messageChan chan papertrailMessage
	syncChan    chan struct{}
	writer      *syslog.Writer
	sentry      sentry
}

type papertrailMessage struct {
	Write   func(writer *syslog.Writer, message string) error
	Message *LogMsg
}

func (papertrail) GetName() string { return "Papertrail" }

func (p papertrail) WriteDebug(message *LogMsg) error {
	p.messageChan <- papertrailMessage{
		Write:   func(w *syslog.Writer, m string) error { return w.Debug(m) },
		Message: message,
	}
	return nil
}

func (p papertrail) WriteInfo(message *LogMsg) error {
	p.messageChan <- papertrailMessage{
		Write:   func(w *syslog.Writer, m string) error { return w.Info(m) },
		Message: message,
	}
	return nil
}

func (p papertrail) WriteWarn(message *LogMsg) error {
	p.messageChan <- papertrailMessage{
		Write:   func(w *syslog.Writer, m string) error { return w.Warning(m) },
		Message: message,
	}
	return nil
}

func (p papertrail) WriteError(message *LogMsg) error {
	p.messageChan <- papertrailMessage{
		Write:   func(w *syslog.Writer, m string) error { return w.Err(m) },
		Message: message,
	}
	return nil
}

func (p papertrail) WritePanic(message *LogMsg) error {
	return p.writer.Emerg(
		fmt.Sprintf(
			"%s: %s %s",
			message.GetLevel(),
			message.GetMessage(),
			message.ConvertAttributesToJSON()))
}

func (p papertrail) runWriter() {
	sequenceNumber := 0
	for {
		message, isOpen := <-p.messageChan
		if !isOpen {
			break
		}

		if message.Write == nil {
			p.syncChan <- struct{}{}
			continue
		}

		sequenceNumber++

		err := message.Write(
			p.writer,
			message.Message.GetMessage()+
				" "+string(message.Message.ConvertAttributesToJSON())+
				" "+strconv.Itoa(sequenceNumber))
		if err != nil {
			log.Printf(
				`Error: Failed to write log %q record: %v`,
				p.GetName(),
				err)
			p.sentry.CaptureMessage(
				NewLogMsg(`failed to write log %q`, p.GetName()).AddErr(err))
		}
	}
}

func (p papertrail) Sync() error {
	p.messageChan <- papertrailMessage{}
	<-p.syncChan
	return nil
}
