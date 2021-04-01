// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package db

import (
	"github.com/palchukovsky/ss"
	"github.com/palchukovsky/ss/ddb"
)

////////////////////////////////////////////////////////////////////////////////

type UserRecord struct{}

func (UserRecord) GetTable() string             { return "User" }
func (UserRecord) GetKeyPartitionField() string { return "id" }
func (UserRecord) GetKeySortField() string      { return "" }

////////////////////////////////////////////////////////////////////////////////

// UserKey is an object ot looks for user by primary key.
type UserKey struct {
	UserRecord
	id ss.UserID
}

// NewUserKey creates new user key instance.
func NewUserKey(id ss.UserID) UserKey { return UserKey{id: id} }

// GetKey returns user key data.
func (key UserKey) GetKey() interface{} {
	return struct {
		ID ss.UserID `json:"id"`
	}{ID: key.id}
}

////////////////////////////////////////////////////////////////////////////////

// User describes user record fields.
type User struct {
	UserRecord
	ID           ss.UserID `json:"id"`
	FirebaseID   string    `json:"fId"`
	CreationTime ddb.Time  `json:"created"`
	// SpotSubscriptionsVersion increases each time when the user subscribed
	// for an event sot or unsubscribed from an event spot.
	SpotSubscriptionsVersion ddb.SubscriptionVersion `json:"spotsVer"`
	Name                     string                  `json:"name"`
	Email                    string                  `json:"email,omitempty"`
	PhoneNumber              string                  `json:"phone,omitempty"`
	PhotoURL                 string                  `json:"photoUrl,omitempty"`
}

// NewUser generates new user record.
func NewFirebaseUser(firebaseID string, name string) (User, UserUniqueIndex) {
	record := User{
		ID:                       ss.NewUserID(),
		FirebaseID:               firebaseID,
		CreationTime:             ddb.Now(),
		SpotSubscriptionsVersion: ddb.NewSubscriptionVersion(1),
		Name:                     name,
	}

	uniqueIndex := UserUniqueIndex{
		Value:  []byte("f#" + record.FirebaseID),
		UserID: record.ID,
	}

	return record, uniqueIndex
}

func (record User) GetData() interface{} { return record }

////////////////////////////////////////////////////////////////////////////////
type UserUniqueIndex struct {
	UserRecord
	Value  []byte    `json:"id"`
	UserID ss.UserID `json:"user"`
}

func (record UserUniqueIndex) GetData() interface{} { return record }

////////////////////////////////////////////////////////////////////////////////

type UserExternalFirabaseIDIndex struct{ UserRecord }

func (UserExternalFirabaseIDIndex) GetIndex() string { return "FirebaseId" }

func (UserExternalFirabaseIDIndex) GetIndexPartitionField() string {
	return "fId"
}

func (UserExternalFirabaseIDIndex) GetIndexSortField() string { return "" }

func (UserExternalFirabaseIDIndex) GetProjection() []string {
	return []string{}
}

////////////////////////////////////////////////////////////////////////////////
