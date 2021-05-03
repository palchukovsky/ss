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

func newPapertrailLogIfSet(
	projectPackage,
	module string,
	config Config,
	sentry sentry,
) (logDestination, error) {

	if config.SS.Log.Papertrail == "" {
		return nil, nil
	}

	result := papertrailLog{
		messageChan: make(chan papertrailLogMessage, 3),
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

	return &result, nil
}

type papertrailLog struct {
	messageChan chan papertrailLogMessage
	syncChan    chan struct{}
	writer      *syslog.Writer
	sentry      sentry
}

type papertrailLogMessage struct {
	Write   func(w *syslog.Writer, m string) error
	Message string
	Type    string
}

func (papertrailLog) GetName() string { return "Papertrail" }

func (pl papertrailLog) WriteDebug(message string) error {
	pl.messageChan <- papertrailLogMessage{
		Write:   func(w *syslog.Writer, m string) error { return w.Debug(m) },
		Message: message,
		Type:    "Debug",
	}
	return nil
}

func (pl papertrailLog) WriteInfo(message string) error {
	pl.messageChan <- papertrailLogMessage{
		Write:   func(w *syslog.Writer, m string) error { return w.Info(m) },
		Message: message,
		Type:    "Info",
	}
	return nil
}

func (pl papertrailLog) WriteWarn(message string) error {
	pl.messageChan <- papertrailLogMessage{
		Write:   func(w *syslog.Writer, m string) error { return w.Warning(m) },
		Message: message,
		Type:    "Warn",
	}
	return nil
}

func (pl papertrailLog) WriteError(message string) error {
	pl.messageChan <- papertrailLogMessage{
		Write:   func(w *syslog.Writer, m string) error { return w.Err(m) },
		Message: message,
		Type:    "Error",
	}
	return nil
}

func (pl *papertrailLog) WritePanic(message string) error {
	return pl.writer.Emerg(message)
}

func (pl papertrailLog) runWriter() {
	sequenceNumber := 0
	for {
		message, isOpen := <-pl.messageChan
		if !isOpen {
			return
		}

		if message.Write == nil {
			pl.syncChan <- struct{}{}
			continue
		}

		sequenceNumber++

		err := message.Write(
			pl.writer,
			message.Message+" "+strconv.Itoa(sequenceNumber))
		if err != nil {
			errMessage := fmt.Sprintf(
				`Error: Failed to write log record: %v`,
				err)
			log.Println(errMessage)
			if err != pl.sentry.CaptureMessage(errMessage) {
				log.Printf(
					"Failed to capture message about %s by Sentry: %v",
					pl.GetName(),
					err)
			}
		}
	}
}

func (pl papertrailLog) Sync() error {
	pl.messageChan <- papertrailLogMessage{}
	<-pl.syncChan
	return nil
}
