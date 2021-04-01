// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package apiauth

import (
	"strconv"
	"time"

	"github.com/palchukovsky/ss"
)

// AccessToken is a access token.
type AccessToken string

// NewAccessToken creates new access token.
func NewAccessToken(user ss.UserID, expirationTime int64,
) (AccessToken, error) {
	result, err := newToken(&accessToken{
		User:           user,
		ExpirationTime: strconv.FormatInt(expirationTime, 16),
	})
	return AccessToken(result), err
}

// ParseAccessToken parses token and verifies its signature. If access token
// is the valid returns user ID and tokening expiration time.
// Otherwise returns pair "user error" and "error of working".
func ParseAccessToken(source string) (ss.UserID, time.Time, error, error) {
	token := accessToken{}
	if userErr, err := parseToken(source, &token); err != nil || userErr != nil {
		return token.User, time.Time{}, userErr, err
	}
	expirationTime, err := parseTimeToken(token.ExpirationTime)
	if err != nil {
		return token.User, time.Time{}, nil, err
	}
	return token.User, expirationTime, nil, nil
}

type accessToken struct {
	User           ss.UserID `json:"u"`
	ExpirationTime string    `json:"e"`
}
