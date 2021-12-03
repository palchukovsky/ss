// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ss

////////////////////////////////////////////////////////////////////////////////

type DeviceID string

func NewDeviceID(source string) DeviceID { return DeviceID(source) }

func (id DeviceID) MarshalLogMsg(destination map[string]interface{}) {
	MarshalLogMsgAttrDump(id, destination)
}

////////////////////////////////////////////////////////////////////////////////

type FirebaseCloudMessagingToken string

func NewFirebaseCloudMessagingToken(source string) FirebaseCloudMessagingToken {
	return FirebaseCloudMessagingToken(source)
}

func (token FirebaseCloudMessagingToken) String() string {
	return string(token)
}

func (token FirebaseCloudMessagingToken) MarshalLogMsg(
	destination map[string]interface{},
) {
	MarshalLogMsgAttrDump(token, destination)
}

////////////////////////////////////////////////////////////////////////////////
