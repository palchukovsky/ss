// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ss

import (
	"log"

	logglylib "github.com/segmentio/go-loggly"
)

func newLogglyIfSet(
	projectPackage,
	module string,
	config Config,
	sentry sentry,
) (logDestination, error) {

	if config.SS.Log.Loggly == "" {
		return nil, nil
	}

	result := loggly{
		messageChan: make(chan logglyMessage, 10),
		syncChan:    make(chan struct{}),
		client:      logglylib.New(config.SS.Log.Loggly),
		sentry:      sentry,
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
	}

	go result.runWriter()

	return result, nil
}

type loggly struct {
	messageChan chan logglyMessage
	syncChan    chan struct{}
	client      *logglylib.Client
	sentry      sentry
	statics     map[string]interface{}
}

type logglyMessage struct {
	Message *LogMsg
	Write   func(client *logglylib.Client, message string) error
}

func (loggly) GetName() string { return "Loggly" }

func (l loggly) WriteDebug(message *LogMsg) error {
	l.write(
		message,
		func(c *logglylib.Client, m string) error { return c.Debug(m) })
	return nil
}

func (l loggly) WriteInfo(message *LogMsg) error {
	l.write(
		message,
		func(c *logglylib.Client, m string) error { return c.Info(m) })
	return nil
}

func (l loggly) WriteWarn(message *LogMsg) error {
	l.write(
		message,
		func(c *logglylib.Client, m string) error { return c.Warn(m) })
	return nil
}

func (l loggly) WriteError(message *LogMsg) error {
	l.write(
		message,
		func(c *logglylib.Client, m string) error { return c.Error(m) })
	return nil
}

func (l loggly) WritePanic(message *LogMsg) error {
	l.write(
		message,
		func(c *logglylib.Client, m string) error { return c.Critical(m) })
	return nil
}

func (l loggly) Sync() error {
	l.messageChan <- logglyMessage{}
	<-l.syncChan
	return nil
}

func (l loggly) write(
	message *LogMsg,
	write func(client *logglylib.Client, message string) error,
) {
	l.messageChan <- logglyMessage{
		Message: message,
		Write:   write,
	}
}

func (l loggly) runWriter() {
	defer func() {
		if err := l.client.Flush(); err != nil {
			log.Printf(
				`Error: Failed to flush log %q record: %v`, l.GetName(), err)
			l.sentry.CaptureMessage(
				NewLogMsg(`failed to flush log %q record`, l.GetName()).AddErr(err))
		}
	}()

	sequenceNumber := 0
	for {
		message, isOpen := <-l.messageChan
		if !isOpen {
			break
		}

		if message.Message == nil {
			if err := l.client.Flush(); err != nil {
				log.Printf(`Error: Failed to sync log %q record: %v`, l.GetName(), err)
				l.sentry.CaptureMessage(
					NewLogMsg(`failed to sync log %q record`, l.GetName()).AddErr(err))
			}
			l.syncChan <- struct{}{}
			continue
		}

		sequenceNumber++
		message.Message.AddVal("n", sequenceNumber)

		for k, v := range l.statics {
			message.Message.AddVal(k, v)
		}

		err := message.Write(
			l.client,
			string(message.Message.ConvertToJSON()))
		if err != nil {
			log.Printf(`Error: Failed to write log %q record: %v`, l.GetName(), err)
			l.sentry.CaptureMessage(
				NewLogMsg(`failed to write log %q record`, l.GetName()).AddErr(err))
		}
	}
}
