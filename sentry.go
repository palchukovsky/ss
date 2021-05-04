// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ss

import (
	"log"
	"time"

	sentryclient "github.com/getsentry/sentry-go"
)

type sentry interface {
	CaptureMessage(string)
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

func (sentryDummy) CaptureMessage(message string)  {}
func (sentryDummy) CaptureException(err error)     {}
func (sentryDummy) Recover(panicValue interface{}) {}
func (sentryDummy) Flush()                         {}

////////////////////////////////////////////////////////////////////////////////

type sentryConnect struct{}

func (sentryConnect) CaptureMessage(message string) {
	if sentryclient.CaptureMessage(message) == nil {
		log.Println("Failed to capture message by Sentry.")
	}
}

func (sentryConnect) CaptureException(err error) {
	if sentryclient.CaptureException(err) == nil {
		log.Println(`Failed to capture exception by Sentry.`)
	}
}

func (sentryConnect) Recover(panicValue interface{}) {
	if sentryclient.CurrentHub().Recover(panicValue) == nil {
		log.Println(`Failed to recover panic by Sentry.`)
	}
}

func (sentryConnect) Flush() {
	if !sentryclient.Flush(2750 * time.Millisecond) {
		log.Println("Not all Sentry records were flushed, timeout was reached.")
	}
}

////////////////////////////////////////////////////////////////////////////////
