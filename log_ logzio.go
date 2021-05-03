// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ss

import (
	"fmt"
	"log"

	lib "github.com/logzio/logzio-go"
)

func newLogzioIfSet(
	projectPackage,
	module string,
	config Config,
	sentry sentry,
) (logDestination, error) {

	if config.SS.Log.Logzio == nil {
		return nil, nil
	}

	result := logzio{
		messageChan: make(chan logzioMessage, 10),
		syncChan:    make(chan struct{}),
		sentry:      sentry,
	}

	var err error
	result.sender, err = lib.New(
		config.SS.Log.Logzio.Token,
		lib.SetUrl(config.SS.Log.Logzio.URL),
	)
	if err != nil {
		panic(err)
	}

	go result.runWriter()

	return result, nil
}

type logzio struct {
	messageChan chan logzioMessage
	syncChan    chan struct{}
	sender      *lib.LogzioSender
	sentry      sentry
}

type logzioMessage struct {
	Message string
	Tag     string
}

func (logzio) GetName() string { return "Logz.io" }

func (l logzio) WriteDebug(message string) error {
	l.write("debug", message)
	return nil
}

func (l logzio) WriteInfo(message string) error {
	l.write("info", message)
	return nil
}

func (l logzio) WriteWarn(message string) error {
	l.write("warn", message)
	return nil
}

func (l logzio) WriteError(message string) error {
	l.write("error", message)
	return nil
}

func (l logzio) WritePanic(message string) error {
	l.write("panic", message)
	return nil
}

func (l logzio) Sync() error {
	l.messageChan <- logzioMessage{}
	<-l.syncChan
	return nil
}

func (l logzio) write(tag, message string) {
	l.messageChan <- logzioMessage{
		Message: message,
		Tag:     tag,
	}
}

func (l logzio) runWriter() {
	defer l.sender.Stop()

	sequenceNumber := 0
	for {
		message, isOpen := <-l.messageChan
		if !isOpen {
			return
		}

		if len(message.Tag) == 0 {
			if err := l.sender.Sync(); err != nil {
				errMessage := fmt.Sprintf(
					`Error: Failed to sync log %q: %v`,
					l.GetName(),
					err)
				log.Println(errMessage)
				if err != l.sentry.CaptureMessage(errMessage) {
					log.Printf(
						"Failed to capture message sync error %q by Sentry: %v",
						l.GetName(),
						err)
				}
			}
			l.syncChan <- struct{}{}
			continue
		}

		sequenceNumber++

		err := l.sender.Send([]byte(
			fmt.Sprintf(
				`{%q:%q,"n":%d}`,
				message.Tag,
				message,
				sequenceNumber)))
		if err != nil {
			errMessage := fmt.Sprintf(
				`Error: Failed to write log %q record: %v`,
				l.GetName(),
				err)
			log.Println(errMessage)
			if err != l.sentry.CaptureMessage(errMessage) {
				log.Printf(
					"Failed to capture message about %q by Sentry: %v",
					l.GetName(),
					err)
			}
		}
	}
}
