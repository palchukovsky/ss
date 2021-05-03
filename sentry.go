// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ss

import (
	"fmt"
	"log"
	"time"

	sentryclient "github.com/getsentry/sentry-go"
)

type sentry interface {
	CaptureMessage(string) error
	CaptureException(error)
	Recover(interface{})
	Flush()
}

func newSentry(
	projectPackage string,
	module string,
	config Config,
) (sentry, error) {

	if config.SS.Log.Sentry == "" {
		return sentryDummy{}, nil
	}

	environment := "production"
	if !config.SS.Build.IsProd() {
		environment = "development"
	}

	err := sentryclient.Init(
		sentryclient.ClientOptions{
			Dsn:              config.SS.Log.Sentry,
			AttachStacktrace: true,
			Release:          config.SS.Build.Version,
			Environment:      environment,
			BeforeSend: func(
				event *sentryclient.Event,
				hint *sentryclient.EventHint,
			) *sentryclient.Event {
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
		return nil, err
	}

	return sentryConnect{}, nil
}

////////////////////////////////////////////////////////////////////////////////

type sentryDummy struct{}

func (sentryDummy) CaptureMessage(message string) error { return nil }
func (sentryDummy) CaptureException(err error)          {}
func (sentryDummy) Recover(panicValue interface{})      {}
func (sentryDummy) Flush()                              {}

////////////////////////////////////////////////////////////////////////////////

type sentryConnect struct{}

func (sentryConnect) CaptureMessage(message string) error {
	if sentryclient.CaptureMessage(message) == nil {
		return fmt.Errorf(
			"Failed to capture message by Sentry. Message: %s",
			message)
	}
	return nil
}

func (sentryConnect) CaptureException(err error) {
	if sentryclient.CaptureException(err) == nil {
		log.Println(`Failed to capture exception by Sentry.`)
	}
}

func (sentryConnect) Recover(panicValue interface{}) {
	sentryclient.CurrentHub().Recover(panicValue)
}

func (sentryConnect) Flush() {
	if !sentryclient.Flush(2 * time.Second) {
		log.Println("Not all Sentry records were flushed, timeout was reached.")
	}
}

////////////////////////////////////////////////////////////////////////////////
