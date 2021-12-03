// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package lib

import (
	"github.com/palchukovsky/ss"
	"github.com/palchukovsky/ss/db"
)

type DeviceUserIndex struct {
	db.DeviceUserIndex
	FCMToken ss.FirebaseCloudMessagingToken `json:"fcm"`
	Key      db.DeviceCryptoKey             `json:"key"`
}

func (r *DeviceUserIndex) Clear() { *r = DeviceUserIndex{} }
