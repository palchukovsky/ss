// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package db

import (
	"time"

	"github.com/palchukovsky/ss"
	"github.com/palchukovsky/ss/ddb"
)

////////////////////////////////////////////////////////////////////////////////

type UserRecord struct{}

func (UserRecord) GetTable() string             { return "User" }
func (UserRecord) GetKeyPartitionField() string { return "id" }
func (UserRecord) GetKeySortField() string      { return "" }

////////////////////////////////////////////////////////////////////////////////
type userKeyValue struct {
	ID ss.UserID `json:"id"`
}

func newUserKeyValueValue(id ss.UserID) userKeyValue {
	return userKeyValue{ID: id}
}

// UserKey is an object ot looks for user by primary key.
type UserKey struct {
	UserRecord
	userKeyValue
}

// NewUserKey creates new user key instance.
func NewUserKey(id ss.UserID) UserKey {
	return UserKey{userKeyValue: newUserKeyValueValue(id)}
}

// GetKey returns user key data.
func (key UserKey) GetKey() interface{} { return key.userKeyValue }

////////////////////////////////////////////////////////////////////////////////

func NewUserAnonymousRecordExpirationTime(start ss.Time) *ss.Time {
	result := start.Add(((time.Hour * 24) * 365) * 292)
	return &result
}

// User describes user record fields.
type User struct {
	UserRecord
	userKeyValue
	FirebaseID   string  `json:"fId"`
	CreationTime ss.Time `json:"created"`
	// SpotMembershipVersion increases each time when the user joins an event
	// spot or leaves it.
	SpotMembershipVersion ddb.MembershipVersion `json:"spotsVer"`
	// OriginalName is the name from the user record source (like Firbase).
	OriginalName string `json:"origName"`
	// OwnName is the name that user set by the app.
	OwnName                       string   `json:"ownName,omitempty"`
	Email                         string   `json:"email,omitempty"`
	PhoneNumber                   string   `json:"phone,omitempty"`
	PhotoURL                      string   `json:"photoUrl,omitempty"`
	AnonymousRecordExpirationTime *ss.Time `json:"anonymExpiration,omitempty"`
}

// NewUser generates new user record.
func NewFirebaseUser(
	firebaseID string,
	name string,
	isAnonymous bool,
) (User, UserUniqueIndex) {
	result := User{
		userKeyValue:          newUserKeyValueValue(ss.NewUserID()),
		FirebaseID:            firebaseID,
		CreationTime:          ss.Now(),
		SpotMembershipVersion: ddb.NewMembershipVersion(1),
		OriginalName:          name,
	}

	if isAnonymous {
		result.AnonymousRecordExpirationTime = NewUserAnonymousRecordExpirationTime(
			result.CreationTime)
	}

	uniqueIndex := UserUniqueIndex{
		Value:  []byte("f#" + result.FirebaseID),
		UserID: result.ID,
	}

	return result, uniqueIndex
}

func (record User) IsAnonymous() bool {
	return record.AnonymousRecordExpirationTime != nil
}

func (record User) GetName() string {
	if record.OwnName != "" {
		return record.OwnName
	}
	return record.OriginalName
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
