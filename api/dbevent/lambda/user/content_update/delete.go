// Copyright 2021-2022, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package usercontentupdatelambda

import (
	"sync"

	"github.com/palchukovsky/ss"
	"github.com/palchukovsky/ss/db"
	"github.com/palchukovsky/ss/ddb"
	dbeventlambda "github.com/palchukovsky/ss/lambda/dbevent"
)

func newDeleter(
	user ss.UserID,
	request dbeventlambda.Request,
	db ddb.Client,
	barrier *sync.WaitGroup,
) Updater {
	return &deleter{
		user:    user,
		db:      db,
		request: request,
		barrier: barrier,
	}
}

////////////////////////////////////////////////////////////////////////////////

type deleter struct {
	user    ss.UserID
	db      ddb.Client
	request dbeventlambda.Request
	barrier *sync.WaitGroup
}

////////////////////////////////////////////////////////////////////////////////

func (deleter *deleter) SpawnUpdate() {
	deleter.barrier.Add(1)
	go func() {
		defer deleter.barrier.Done()
		defer func() { ss.S.Log().CheckExit(recover()) }()

		deleter.deleteConnections()
	}()

	deleter.barrier.Add(1)
	go func() {
		defer deleter.barrier.Done()
		defer func() { ss.S.Log().CheckExit(recover()) }()

		deleter.deleteDevices()
	}()
}

////////////////////////////////////////////////////////////////////////////////

type connectionRecord struct {
	db.ConnectionUserIndex
	db.ConnectionKeyValue
}

func (record *connectionRecord) Clear() { *record = connectionRecord{} }

func (deleter *deleter) deleteConnections() {

	var record connectionRecord
	it := deleter.
		db.
		Index(&record).
		Query("user = :u", ddb.Values{":u": deleter.user}).
		RequestPaged()

	for it.Next() {
		deleter.barrier.Add(1)
		go func(key db.ConnectionKeyValue) {
			defer deleter.barrier.Done()
			defer func() { ss.S.Log().CheckExit(recover()) }()

			trans := deleter.db.Delete(db.NewConnectionKey(key.ID))
			trans.AllowConditionalCheckFail()

			if trans.Request().IsSuccess() {
				deleter.request.Log().Info(
					ss.NewLogMsg("deleted connection").Add(key.ID))
			} else {
				deleter.request.Log().Debug(
					ss.NewLogMsg("no connection record found").Add(key.ID))
			}

		}(record.ConnectionKeyValue)

	}

}

////////////////////////////////////////////////////////////////////////////////

type deviceRecord struct {
	db.DeviceUserIndex
	db.DeviceKeyValue
}

func (record *deviceRecord) Clear() { *record = deviceRecord{} }

func (deleter *deleter) deleteDevices() {
	deleter.barrier.Add(1)
	go func() {
		defer deleter.barrier.Done()
		defer func() { ss.S.Log().CheckExit(recover()) }()

		var record deviceRecord
		it := deleter.
			db.
			Index(&record).
			Query("user = :u", ddb.Values{":u": deleter.user}).
			RequestPaged()

		for it.Next() {
			deleter.barrier.Add(1)
			go func(key db.DeviceKeyValue) {
				defer deleter.barrier.Done()
				defer func() { ss.S.Log().CheckExit(recover()) }()

				trans := deleter.db.Delete(db.NewDeviceKey(key.FCMToken))
				trans.AllowConditionalCheckFail()

				if trans.Request().IsSuccess() {
					deleter.request.Log().Info(
						ss.NewLogMsg("deleted device").Add(key.FCMToken))
				} else {
					deleter.request.Log().Debug(
						ss.NewLogMsg("no device record found").Add(key.FCMToken))
				}

			}(record.DeviceKeyValue)
		}
	}()
}

////////////////////////////////////////////////////////////////////////////////
