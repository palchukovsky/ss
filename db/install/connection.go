// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package dbinstall

import (
	"github.com/palchukovsky/ss/db"
	"github.com/palchukovsky/ss/ddb"
	ddbinstall "github.com/palchukovsky/ss/ddb/install"
)

type connection struct{ ddbinstall.TableAbstraction }

func newConnectionTable(ddb ddbinstall.DB) ddbinstall.Table {
	return connection{
		TableAbstraction: ddbinstall.NewTableAbstraction(ddb, db.Connection{}),
	}
}

func (table connection) Create() error {
	return table.TableAbstraction.Create(
		[]ddb.IndexRecord{&db.ConnectionIDByUser{}})
}

func (table connection) Setup() error {
	err := table.EnableStreams(
		ddbinstall.StreamViewTypeNone,
		[]ddbinstall.Stream{
			ddbinstall.NewStream("Init"),
		})
	if err != nil {
		return err
	}
	return table.EnableTimeToLive("expired")
}
