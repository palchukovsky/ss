// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

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
	Write   func(w *syslog.Writer, m string) error
	Message string
	Type    string
}

func (papertrail) GetName() string { return "Papertrail" }

func (p papertrail) WriteDebug(message string) error {
	p.messageChan <- papertrailMessage{
		Write:   func(w *syslog.Writer, m string) error { return w.Debug(m) },
		Message: message,
		Type:    "Debug",
	}
	return nil
}

func (p papertrail) WriteInfo(message string) error {
	p.messageChan <- papertrailMessage{
		Write:   func(w *syslog.Writer, m string) error { return w.Info(m) },
		Message: message,
		Type:    "Info",
	}
	return nil
}

func (p papertrail) WriteWarn(message string) error {
	p.messageChan <- papertrailMessage{
		Write:   func(w *syslog.Writer, m string) error { return w.Warning(m) },
		Message: message,
		Type:    "Warn",
	}
	return nil
}

func (p papertrail) WriteError(message string) error {
	p.messageChan <- papertrailMessage{
		Write:   func(w *syslog.Writer, m string) error { return w.Err(m) },
		Message: message,
		Type:    "Error",
	}
	return nil
}

func (p papertrail) WritePanic(message string) error {
	return p.writer.Emerg(message)
}

func (p papertrail) runWriter() {
	sequenceNumber := 0
	for {
		message, isOpen := <-p.messageChan
		if !isOpen {
			return
		}

		if message.Write == nil {
			p.syncChan <- struct{}{}
			continue
		}

		sequenceNumber++

		err := message.Write(
			p.writer,
			message.Message+" "+strconv.Itoa(sequenceNumber))
		if err != nil {
			errMessage := fmt.Sprintf(
				`Error: Failed to write log %q record: %v`,
				p.GetName(),
				err)
			log.Println(errMessage)
			if err != p.sentry.CaptureMessage(errMessage) {
				log.Printf(
					"Failed to capture message about %q by Sentry: %v",
					p.GetName(),
					err)
			}
		}
	}
}

func (p papertrail) Sync() error {
	p.messageChan <- papertrailMessage{}
	<-p.syncChan
	return nil
}
