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
		messageChan: make(chan *LogMsg, 10),
		syncChan:    make(chan struct{}),
		client:      logglylib.New(config.SS.Log.Loggly),
		sentry:      sentry,
	}

	go result.runWriter()

	return result, nil
}

type loggly struct {
	messageChan chan *LogMsg
	syncChan    chan struct{}
	client      *logglylib.Client
	sentry      sentry
}

func (loggly) GetName() string { return "Loggly" }

func (l loggly) WriteDebug(message *LogMsg) error {
	l.write(message)
	return nil
}

func (l loggly) WriteInfo(message *LogMsg) error {
	l.write(message)
	return nil
}

func (l loggly) WriteWarn(message *LogMsg) error {
	l.write(message)
	return nil
}

func (l loggly) WriteError(message *LogMsg) error {
	l.write(message)
	return nil
}

func (l loggly) WritePanic(message *LogMsg) error {
	l.write(message)
	return nil
}

func (l loggly) Sync() error {
	l.messageChan <- nil
	<-l.syncChan
	return nil
}

func (l loggly) write(message *LogMsg) { l.messageChan <- message }

func (l loggly) runWriter() {
	defer func() {
		if err := l.client.Flush(); err != nil {
			log.Printf(
				`Error: Failed to flush log %q record: %v`, l.GetName(), err)
			l.sentry.CaptureMessage(
				NewLogMsg(`failed to flush log %q record`, l.GetName()).AddErr(err))
		}
	}()

	for {
		message, isOpen := <-l.messageChan
		if !isOpen {
			break
		}

		if message == nil {
			if err := l.client.Flush(); err != nil {
				log.Printf(`Error: Failed to sync log %q record: %v`, l.GetName(), err)
				l.sentry.CaptureMessage(
					NewLogMsg(`failed to sync log %q record`, l.GetName()).AddErr(err))
			}
			l.syncChan <- struct{}{}
			continue
		}

		if err := l.client.Send(message.MarshalMap()); err != nil {
			log.Printf(`Error: Failed to write log %q record: %v`, l.GetName(), err)
			l.sentry.CaptureMessage(
				NewLogMsg(`failed to write log %q record`, l.GetName()).AddErr(err))
		}
	}
}
