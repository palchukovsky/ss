// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package db

import (
	"github.com/palchukovsky/ss"
	"github.com/palchukovsky/ss/crypto"
)

////////////////////////////////////////////////////////////////////////////////

type DeviceRecord struct{}

func (DeviceRecord) GetTable() string             { return "Device" }
func (DeviceRecord) GetKeyPartitionField() string { return "fcm" }
func (DeviceRecord) GetKeySortField() string      { return "" }

////////////////////////////////////////////////////////////////////////////////

type deviceKeyValue struct {
	FCMToken ss.FirebaseCloudMessagingToken `json:"fcm"`
}

func newDeviceKeyValue(fcmToken ss.FirebaseCloudMessagingToken) deviceKeyValue {
	return deviceKeyValue{FCMToken: fcmToken}
}

type DeviceKey struct {
	DeviceRecord
	deviceKeyValue
}

func NewDeviceKey(fcmToken ss.FirebaseCloudMessagingToken) DeviceKey {
	return DeviceKey{
		deviceKeyValue: newDeviceKeyValue(fcmToken),
	}
}

func (key DeviceKey) GetKey() interface{} { return key.deviceKeyValue }

////////////////////////////////////////////////////////////////////////////////

type DeviceCryptoKey = crypto.AES128Key

type Device struct {
	DeviceRecord
	deviceKeyValue
	ID   ss.DeviceID     `json:"id"`
	User ss.UserID       `json:"user"`
	Key  DeviceCryptoKey `json:"key"`
}

func NewDevice(
	id ss.DeviceID,
	fcmToken ss.FirebaseCloudMessagingToken,
	user ss.UserID,
	key DeviceCryptoKey,
) Device {
	return Device{
		deviceKeyValue: newDeviceKeyValue(fcmToken),
		ID:             id,
		User:           user,
		Key:            key,
	}
}

func (record Device) GetData() interface{} { return record }
func (record *Device) Clear()              { *record = Device{} }

func (record Device) MarshalLogMsg(destination map[string]interface{}) {
	ss.MarshalLogMsgAttrDump(record, destination)
}

////////////////////////////////////////////////////////////////////////////////

type DeviceUserIndex struct{ DeviceRecord }

// GetIndex returns index name.
func (DeviceUserIndex) GetIndex() string { return "User" }

// GetIndexPartitionField returns index partition field name.
func (DeviceUserIndex) GetIndexPartitionField() string { return "user" }

// GetIndexSortField returns index sort field name.
func (DeviceUserIndex) GetIndexSortField() string { return "" }

func (DeviceUserIndex) GetProjection() []string { return []string{} }

////////////////////////////////////////////////////////////////////////////////
