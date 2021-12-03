// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package gatewaylambda

import (
	"errors"
	"fmt"

	"github.com/palchukovsky/ss"
	"github.com/palchukovsky/ss/crypto"
	"github.com/palchukovsky/ss/db"
)

////////////////////////////////////////////////////////////////////////////////

type ClientInfo struct {
	Version  string                          `json:"ver"`
	Device   ss.DeviceID                     `json:"device"`
	FCMToken *ss.FirebaseCloudMessagingToken `json:"fcm,omitempty"`
	Key      db.DeviceCryptoKey              `json:"key"`
}

func NewClientInfo(
	getHeader func(name string) (string, bool),
) (
	result ClientInfo,
	err error,
) {

	if result.Version, err = NewClientVersion(getHeader); err != nil {
		return result, fmt.Errorf(`failed to parse client version: "%w"`, err)
	}

	if result.Key, err = NewClientKey(getHeader); err != nil {
		return result, fmt.Errorf(`failed to parse client key: "%w"`, err)
	}

	// Headers have to be in lowercase for better compression.
	// Also, Cloudflare converts it in lower case.

	headerPrefix := ss.S.Config().HeaderPrefix

	{
		var val string
		{
			var has bool
			if val, has = getHeader(headerPrefix + "-device"); !has || val == "" {
				return result, errors.New("client didn't provide its device ID")
			}
		}
		if len(val) > 1024 { // it also controlled somewhere in schema
			return result, errors.New("client provided too long device ID")
		}
		result.Device = ss.NewDeviceID(val)
	}

	if val, has := getHeader(headerPrefix + "-fcm"); has {
		if len(val) < 1 || len(val) > 4096 { // it also controlled somewhere in schema
			return result, errors.New("client provided too long FCM token")
		}
		fcmToken := ss.NewFirebaseCloudMessagingToken(val)
		result.FCMToken = &fcmToken
	}

	return result, nil
}

func (info ClientInfo) MarshalLogMsg(destination map[string]interface{}) {
	ss.MarshalLogMsgAttrDump(info, destination)
}

////////////////////////////////////////////////////////////////////////////////

func NewClientVersion(
	getHeader func(name string) (string, bool),
) (string, error) {
	// Headers have to be in lowercase for better compression.
	// Also, Cloudflare converts it in lower case.
	result, has := getHeader(ss.S.Config().HeaderPrefix + "-ver")
	if !has || result == "" {
		return result, errors.New("client didn't provide its version")
	}
	if len(result) > 64 {
		return result, errors.New("client didn't provide its version")
	}
	return result, nil
}

////////////////////////////////////////////////////////////////////////////////

func NewClientKey(
	getHeader func(name string) (string, bool),
) (db.DeviceCryptoKey, error) {
	// Headers have to be in lowercase for better compression.
	// Also, Cloudflare converts it in lower case.
	key, has := getHeader(ss.S.Config().HeaderPrefix + "-key")
	if !has {
		return db.DeviceCryptoKey{}, errors.New("client key is not provided")
	}
	result, err := crypto.NewAES128KeyFromBase64(key)
	if err != nil {
		return result, fmt.Errorf(`failed to parse client key: "%w"`, err)
	}
	return result, nil
}

////////////////////////////////////////////////////////////////////////////////
