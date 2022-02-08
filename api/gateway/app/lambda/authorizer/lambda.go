// Copyright 2021-2022, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package authorizerlambda

import (
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/palchukovsky/ss"
	apiauth "github.com/palchukovsky/ss/api/gateway/auth"
)

func Run(initService func(projectPackage string, params ss.ServiceParams)) {

	initService("app", ss.ServiceParams{IsAuth: true})
	defer func() { ss.S.Log().CheckExit(recover()) }()

	config := ss.S.Config()
	policy = events.APIGatewayCustomAuthorizerPolicy{
		Version: "2012-10-17",
		Statement: []events.IAMPolicyStatement{
			{
				Action: []string{"execute-api:Invoke"},
				Effect: "Allow",
				Resource: []string{fmt.Sprintf(
					"arn:aws:execute-api:%s:%s:%s/%s/$connect",
					config.AWS.Region,
					config.AWS.AccountID,
					config.AWS.Gateway.App.ID,
					ss.S.Build().Version)},
			},
			{
				Action: []string{
					"execute-api:InvalidateCache",
					"execute-api:ManageConnections",
				},
				Effect:   "Deny",
				Resource: []string{"*"},
			},
		},
	}

	ss.S.Log().Started()

	lambda.Start(handle)
}

////////////////////////////////////////////////////////////////////////////////

type (
	request  = events.APIGatewayCustomAuthorizerRequestTypeRequest
	response = events.APIGatewayCustomAuthorizerResponse
)

// Headers have to be in lowercase for better compression.
// Also, Cloudflare converts it in lower case.
const authHeaderName = "auth"

var (
	policy events.APIGatewayCustomAuthorizerPolicy
)

func handle(ctx context.Context, request request) (response, error) {
	ss.S.StartLambda(
		func() []ss.LogMsgAttr {
			// Duplicates request data in the logs records with panic,
			// but not in other records.
			return ss.NewLogMsgAttrRequestDumps(request)
		})
	defer func() { ss.S.CompleteLambda(recover()) }()

	accessToken, hasAccessToken := request.Headers[authHeaderName]
	if !hasAccessToken {
		newLog(request).Error(
			ss.NewLogMsg(`no access token`).AddRequest(request))
		return response{}, errors.New("no access token")
	}

	user, expirationTime, userErr, err := apiauth.ParseAccessToken(accessToken)
	if err != nil {
		if ss.S.Build().IsProd() {
			newLog(request).Error(
				ss.
					NewLogMsg(`failed to parse access token`).
					AddErr(err).
					AddRequest(request))
		} else {
			newLog(request).Debug(
				ss.NewLogMsg(`Failed to parse access token`).
					AddErr(err).
					AddRequest(request))
		}
		return response{}, errors.New("failed to parse access token")
	}
	if userErr != nil {
		newLog(request).Warn(
			ss.NewLogMsg(`failed to parse access token`).
				AddErr(userErr).
				AddRequest(request))
		return response{}, errors.New("failed to parse access token")
	}

	if !expirationTime.After(ss.Now()) {
		newUserLog(request, user).Debug(
			ss.NewLogMsg("access key expired at %s", expirationTime))
		// Special return to generate 401 ("Unauthorized" is case sensitive).
		return response{}, errors.New("Unauthorized")

	}

	result := response{
		PolicyDocument: policy,
		PrincipalID:    user.String(),
		Context:        map[string]interface{}{},
	}
	for key, val := range request.Headers {
		result.Context[key] = val
	}
	return result, nil
}

func newLog(request request) ss.LogStream {
	return ss.S.Log().NewSession(
		ss.
			NewLogPrefix(
				func() []ss.LogMsgAttr {
					return ss.NewLogMsgAttrRequestDumps(request)
				}).
			AddRequestID(request.RequestContext.RequestID))
}

func newUserLog(request request, user ss.UserID) ss.LogStream {
	return ss.S.Log().NewSession(
		ss.
			NewLogPrefix(
				func() []ss.LogMsgAttr {
					return ss.NewLogMsgAttrRequestDumps(request)
				}).
			Add(user).
			AddRequestID(request.RequestContext.RequestID))
}
