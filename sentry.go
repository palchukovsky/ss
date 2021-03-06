// Copyright 2021-2022, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ss

import (
	"log"

	sentryclient "github.com/getsentry/sentry-go"
)

type sentry interface {
	CaptureMessage(*LogMsg)
	Recover(*LogMsg)
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

	err := sentryclient.Init(
		sentryclient.ClientOptions{
			Dsn:              config.SS.Log.Sentry,
			AttachStacktrace: true,
			Release:          config.SS.Build.Version,
			Environment:      config.SS.Build.GetEnvironment(),
			BeforeSend: func(
				event *sentryclient.Event,
				hint *sentryclient.EventHint,
			) *sentryclient.Event {
				event.Tags["module"] = module
				event.Tags["package"] = projectPackage
				event.Tags["build"] = config.SS.Build.ID
				event.Tags["commit"] = config.SS.Build.Commit
				event.Tags["builder"] = config.SS.Build.Builder
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

func (sentryDummy) CaptureMessage(*LogMsg) {}
func (sentryDummy) Recover(*LogMsg)        {}
func (sentryDummy) Flush()                 {}

////////////////////////////////////////////////////////////////////////////////

type sentryConnect struct{}

func (s sentryConnect) CaptureMessage(message *LogMsg) {
	if sentryclient.CaptureEvent(s.newEvent(message, false)) == nil {
		log.Println("Failed to capture message by Sentry.")
	}
}

func (s sentryConnect) Recover(message *LogMsg) {

	event := s.newEvent(message, true)

	if sentryclient.CaptureEvent(event) == nil {
		log.Println(`Failed to recover panic by Sentry.`)
	}
}

func (sentryConnect) Flush() {
	if !sentryclient.Flush(S.Config().AWS.LambdaTimeout / 2) {
		log.Println("Not all Sentry records were flushed, timeout was reached.")
	}
}

func (s sentryConnect) newEvent(
	source *LogMsg,
	isCrash bool,
) *sentryclient.Event {
	result := sentryclient.NewEvent()

	switch source.GetLevel() {
	case logLevelDebug:
		result.Level = sentryclient.LevelDebug
	case logLevelInfo:
		result.Level = sentryclient.LevelInfo
	case logLevelWarn:
		result.Level = sentryclient.LevelWarning
	case logLevelError:
		result.Level = sentryclient.LevelError
	default: // logLevelPanic also here
		result.Level = sentryclient.LevelFatal
	}

	result.Threads = []sentryclient.Thread{{
		Stacktrace: sentryclient.NewStacktrace(),
		Crashed:    isCrash,
		Current:    true,
	}}

	result.Extra = source.MarshalAttributesMap()
	if user, has := result.Extra[logMsgNodeUser]; has {
		result.User.ID = user.(UserID).String()
		delete(result.Extra, logMsgNodeUser)
	}

	result.Message = source.GetMessage()

	// Exception will be stored with Sentry record as "dumps".

	return result
}

////////////////////////////////////////////////////////////////////////////////
