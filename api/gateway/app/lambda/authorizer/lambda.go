// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package authorizerlambda

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/palchukovsky/ss"
	apiauth "github.com/palchukovsky/ss/api/gateway/auth"
)

func Run(initService func(projectPackage string, params ss.ServiceParams)) {

	initService("app", ss.ServiceParams{IsAuth: true})
	defer ss.S.Log().CheckExit(recover())

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

type request = events.APIGatewayCustomAuthorizerRequestTypeRequest
type response = events.APIGatewayCustomAuthorizerResponse

const authHeaderName = "Auth"

var authHeaderNameLower = strings.ToLower(authHeaderName)

var policy events.APIGatewayCustomAuthorizerPolicy

func handle(ctx context.Context, request request) (response, error) {
	defer ss.S.Log().CheckExit(recover())

	accessToken, hasAccessToken := request.Headers[authHeaderName]
	if !hasAccessToken {
		accessToken, hasAccessToken = request.Headers[authHeaderNameLower]
		if !hasAccessToken {
			getLog(request).Error(`No access token. Dump: %s`, ss.Dump(request))
			return response{}, errors.New("no access token")
		}
	}

	user, expirationTime, userErr, err := apiauth.ParseAccessToken(accessToken)
	if err != nil {
		if ss.S.Build().IsProd() {
			getLog(request).Error(
				`Failed to parse access token: "%v". Dump: %s`,
				err,
				ss.Dump(request))
		} else {
			getLog(request).Debug(
				`Failed to parse access token: "%v". Dump: %s`,
				err,
				ss.Dump(request))
		}
		return response{}, errors.New("failed to parse access token")
	}
	if userErr != nil {
		getLog(request).Warn(`Failed to parse access token: "%v". Dump: %s`,
			userErr, ss.Dump(request))
		return response{}, errors.New("failed to parse access token")
	}

	if !expirationTime.After(ss.Now()) {
		getUserLog(request, user).Debug("Access key expired at %s.", expirationTime)
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

func getLog(request request) ss.ServiceLogStream {
	return ss.S.Log().NewSession(
		fmt.Sprintf("*.%s", request.RequestContext.RequestID))
}

func getUserLog(request request, user ss.UserID) ss.ServiceLogStream {
	return ss.S.Log().NewSession(fmt.Sprintf("%s..%s",
		user, request.RequestContext.RequestID))
}
