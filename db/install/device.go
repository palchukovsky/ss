// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package dbinstall

import (
	"github.com/palchukovsky/ss"
	"github.com/palchukovsky/ss/db"
	"github.com/palchukovsky/ss/ddb"
	ddbinstall "github.com/palchukovsky/ss/ddb/install"
	push "github.com/palchukovsky/ss/push/lib"
)

type device struct{ ddbinstall.TableAbstraction }

func newDeviceTable(ddb ddbinstall.DB, log ss.Log) ddbinstall.Table {
	return device{
		TableAbstraction: ddbinstall.NewTableAbstraction(ddb, db.Device{}, log),
	}
}

func (table device) Create() error {
	return table.TableAbstraction.Create([]ddb.IndexRecord{
		&push.DeviceUserIndex{},
	})
}

func (table device) Setup() error { return nil }
func (device) InsertData() error  { return nil }
