// Copyright 2021-2022, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package db

import (
	"github.com/palchukovsky/ss"
	"github.com/palchukovsky/ss/ddb"
)

type deleteUserRecord struct {
	UserKey

	FirebaseID string `json:"fId"`
}

func (record *deleteUserRecord) Clear() { *record = deleteUserRecord{} }

func DeleteUser(id ss.UserID, db ddb.Client) bool {

	record := deleteUserRecord{UserKey: NewUserKey(id)}
	if !db.Find(&record).Request() {
		return false
	}

	if record.FirebaseID == "" {
		// This version works only with Firebase.
		ss.S.Log().Panic(ss.NewLogMsg("user does not have Firebase ID").Add(id))
	}

	trans := ddb.NewWriteTrans(true)
	recordCondition := trans.Delete(record).AllowConditionalCheckFail()
	trans.Delete(newFirebaseUserUniqueIndex(id, record.FirebaseID))

	if result := db.Write(trans); !result.IsSuccess() {
		if result.ParseConditions().IsPassed(recordCondition) {
			ss.S.Log().Panic(ss.NewLogMsg("user does not have uniquer key").Add(id))
		}
		return false
	}

	return true
}
