// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ss

import (
	"fmt"
	"log"
	"strconv"

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
	}

	go result.runWriter()

	return result, nil
}

type loggly struct {
	messageChan chan logglyMessage
	syncChan    chan struct{}
	client      *logglylib.Client
	sentry      sentry
}

type logglyMessage struct {
	Message string
	Write   func(client *logglylib.Client, message string) error
}

func (loggly) GetName() string { return "Loggly" }

func (l loggly) WriteDebug(message string) error {
	l.write(
		message,
		func(c *logglylib.Client, m string) error { return c.Debug(m) })
	return nil
}

func (l loggly) WriteInfo(message string) error {
	l.write(
		message,
		func(c *logglylib.Client, m string) error { return c.Info(m) })
	return nil
}

func (l loggly) WriteWarn(message string) error {
	l.write(
		message,
		func(c *logglylib.Client, m string) error { return c.Warn(m) })
	return nil
}

func (l loggly) WriteError(message string) error {
	l.write(
		message,
		func(c *logglylib.Client, m string) error { return c.Error(m) })
	return nil
}

func (l loggly) WritePanic(message string) error {
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
	message string,
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
			l.sentry.CaptureException(
				fmt.Errorf(`Failed to flush log %q record: %w`, l.GetName(), err))
		}
	}()

	sequenceNumber := 0
	for {
		message, isOpen := <-l.messageChan
		if !isOpen {
			return
		}

		if message.Write == nil {
			if err := l.client.Flush(); err != nil {
				log.Printf(
					`Error: Failed to sync log %q record: %v`, l.GetName(), err)
				l.sentry.CaptureException(
					fmt.Errorf(`Failed to sync log %q record: %w`, l.GetName(), err))
			}
			l.syncChan <- struct{}{}
			continue
		}

		sequenceNumber++

		err := message.Write(
			l.client,
			message.Message+" "+strconv.Itoa(sequenceNumber))
		if err != nil {
			log.Printf(
				`Error: Failed to write log %q record: %v`, l.GetName(), err)
			l.sentry.CaptureException(
				fmt.Errorf(`Failed to write log %q record: %w`, l.GetName(), err))
		}
	}
}
