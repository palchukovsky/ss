// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package dbinstall

import (
	"github.com/palchukovsky/ss"
	"github.com/palchukovsky/ss/db"
	"github.com/palchukovsky/ss/ddb"
	ddbinstall "github.com/palchukovsky/ss/ddb/install"
)

type connection struct{ ddbinstall.TableAbstraction }

func newConnectionTable(ddb ddbinstall.DB, log ss.Log) ddbinstall.Table {
	return connection{
		TableAbstraction: ddbinstall.NewTableAbstraction(ddb, db.Connection{}, log),
	}
}

func (table connection) Create() error {
	return table.TableAbstraction.Create(
		[]ddb.IndexRecord{&db.ConnectionIDByUser{}})
}

func (table connection) Setup() error {
	err := table.EnableStreams(
		// -------------------------------------------------------------------------
		/*
			Required by BUZZ-78, but disabled to don't send full record until
			version control required (see substring BUZZ-78 for other details):

			ddbinstall.StreamViewTypeNew,
		*/
		ddbinstall.StreamViewTypeNone,
		// -------------------------------------------------------------------------
		[]ddbinstall.Stream{
			ddbinstall.NewStream("Init"),
		})
	if err != nil {
		return err
	}
	return table.EnableTimeToLive("expiration")
}

func (connection) InsertData() error { return nil }
