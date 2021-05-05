// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ss

import (
	"fmt"
	"log"
	"reflect"
	"time"

	sentryclient "github.com/getsentry/sentry-go"
)

type sentry interface {
	CaptureMessage(*LogMsg)
	Recover(interface{}, *LogMsg)
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

func (sentryDummy) CaptureMessage(*LogMsg)       {}
func (sentryDummy) Recover(interface{}, *LogMsg) {}
func (sentryDummy) Flush()                       {}

////////////////////////////////////////////////////////////////////////////////

type sentryConnect struct{}

func (s sentryConnect) CaptureMessage(message *LogMsg) {
	event := s.newEvent(message, false)

	event.Message = message.GetMessage()

	if errs := message.GetErrs(); len(errs) > 0 {
		for i, err := range errs {
			if i == 0 {
				event.Message += ": "
			} else {
				event.Message += "; "
			}
			event.Message += fmt.Sprintf("%v", err.Get())
		}
	}

	if sentryclient.CaptureEvent(event) == nil {
		log.Println("Failed to capture message by Sentry.")
	}
}

func (s sentryConnect) Recover(panicValue interface{}, message *LogMsg) {

	event := s.newEvent(message, true)

	switch err := panicValue.(type) {
	case error:
		const maxErrorDepthFromDentryclient = 10
		for i := 0; i < maxErrorDepthFromDentryclient && err != nil; i++ {
			event.Exception = append(
				event.Exception,
				sentryclient.Exception{
					Value:      err.Error(),
					Type:       reflect.TypeOf(err).String(),
					Stacktrace: sentryclient.ExtractStacktrace(err),
				})
			switch previous := err.(type) {
			case interface{ Unwrap() error }:
				err = previous.Unwrap()
			case interface{ Cause() error }:
				err = previous.Cause()
			default:
				err = nil
			}
		}
	case string:
		event.Message = fmt.Sprintf("%s: %s", message.GetMessage(), err)
	default:
		event.Message = fmt.Sprintf("%s: %#v", message.GetMessage(), err)
	}

	if sentryclient.CaptureEvent(event) == nil {
		log.Println(`Failed to recover panic by Sentry.`)
	}
}

func (sentryConnect) Flush() {
	if !sentryclient.Flush(2750 * time.Millisecond) {
		log.Println("Not all Sentry records were flushed, timeout was reached.")
	}
}

func (sentryConnect) newEvent(
	source *LogMsg,
	isCrash bool,
) *sentryclient.Event {
	result := sentryclient.NewEvent()

	result.Message = source.GetMessage()

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

	return result
}

////////////////////////////////////////////////////////////////////////////////
