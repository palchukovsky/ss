// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ss

import (
	"fmt"
	"log"

	logziolib "github.com/logzio/logzio-go"
)

func newLogzioIfSet(
	projectPackage string,
	module string,
	config Config,
	sentry sentry,
) (logDestination, error) {
	logConfig := config.SS.Log.Logzio
	if logConfig == nil {
		return nil, nil
	}

	result := logzio{
		messageChan: make(chan *LogMsg, 10),
		syncChan:    make(chan struct{}),
		sentry:      sentry,
	}

	var err error
	result.sender, err = logziolib.New(
		logConfig.Token,
		logziolib.SetUrl(fmt.Sprintf("https://%s:8071", logConfig.Host)),
		// logziolib.SetDebug(os.Stderr),
		logziolib.SetTempDirectory("/tmp/logzio_tmp"))
	if err != nil {
		panic(err)
	}

	go result.runWriter()

	return result, nil
}

type logzio struct {
	messageChan chan *LogMsg
	syncChan    chan struct{}
	sender      *logziolib.LogzioSender
	sentry      sentry
}

func (logzio) GetName() string { return "Logz.io" }

func (l logzio) WriteDebug(message *LogMsg) error {
	l.messageChan <- message
	return nil
}

func (l logzio) WriteInfo(message *LogMsg) error {
	l.messageChan <- message
	return nil
}

func (l logzio) WriteWarn(message *LogMsg) error {
	l.messageChan <- message
	return nil
}

func (l logzio) WriteError(message *LogMsg) error {
	l.messageChan <- message
	return nil
}

func (l logzio) WritePanic(message *LogMsg) error {
	l.messageChan <- message
	return nil
}

func (l logzio) Sync() error {
	l.messageChan <- nil
	<-l.syncChan
	return nil
}

func (l logzio) runWriter() {
	defer l.sender.Stop()

	for {
		message, isOpen := <-l.messageChan
		if !isOpen {
			break
		}

		if message == nil {
			if err := l.sender.Sync(); err != nil {
				log.Printf(`Error: Failed to sync log %q record: %v`, l.GetName(), err)
				l.sentry.CaptureMessage(
					NewLogMsg(`failed to sync log %q record`, l.GetName()).AddErr(err))
			}
			l.syncChan <- struct{}{}
			continue
		}

		if err := l.sender.Send(message.ConvertToJSON()); err != nil {
			log.Printf(
				`Error: Failed to write log %q record: %v`,
				l.GetName(),
				err)
			l.sentry.CaptureMessage(
				NewLogMsg(`failed to write log %q record`, l.GetName()).AddErr(err))
		}
	}
	l.sender.Drain()
}
