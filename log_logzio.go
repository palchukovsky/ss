// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ss

import (
	"log"

	logziolib "github.com/logzio/logzio-go"
)

func newLogzioIfSet(
	projectPackage string,
	module string,
	config Config,
	sentry sentry,
) (logDestination, error) {

	if config.SS.Log.Logzio == "" {
		return nil, nil
	}

	result := logzio{
		messageChan: make(chan *LogMsg, 10),
		syncChan:    make(chan struct{}),
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

	var err error
	result.sender, err = logziolib.New(
		config.SS.Log.Logzio,
		// logziolib.SetDebug(os.Stderr),
		logziolib.SetTempDirectory("logzio_tmp"))
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
	statics     map[string]interface{}
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

	sequenceNumber := 0
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

		sequenceNumber++
		message.AddVal("n", sequenceNumber)

		for k, v := range l.statics {
			message.AddVal(k, v)
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
