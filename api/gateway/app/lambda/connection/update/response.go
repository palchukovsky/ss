// Copyright 2021-2022, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package connectionupdatelambda

import "github.com/palchukovsky/ss"

type request struct {
	Device   ss.DeviceID                    `json:"device"`
	FCMToken ss.FirebaseCloudMessagingToken `json:"fcm"`
}

type response struct{}

func newResponse() response { return response{} }
