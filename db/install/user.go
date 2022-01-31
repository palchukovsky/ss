// Copyright 2021-2022, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package dbinstall

import (
	"github.com/palchukovsky/ss"
	lambda "github.com/palchukovsky/ss/api/gateway/auth/lambda"
	"github.com/palchukovsky/ss/db"
	"github.com/palchukovsky/ss/ddb"
	ddbinstall "github.com/palchukovsky/ss/ddb/install"
)

type user struct {
	ddbinstall.TableAbstraction

	streamsFactory func() *ddbinstall.Streams
}

func newUserTable(
	ddb ddbinstall.DB,
	log ss.Log,
	streamsFactory func() *ddbinstall.Streams,
) ddbinstall.Table {
	return user{
		TableAbstraction: ddbinstall.NewTableAbstraction(ddb, db.User{}, log),
		streamsFactory:   streamsFactory,
	}
}

func (table user) Create() error {
	return table.TableAbstraction.Create(
		[]ddb.IndexRecord{&lambda.FirebaseIndex{}})
}

func (table user) Setup() error {
	if err := table.EnableTimeToLive("anonymExpiration"); err != nil {
		return err
	}
	if streams := table.streamsFactory(); streams != nil {
		if err := table.EnableStreams(*streams); err != nil {
			return err
		}
	}
	return nil
}

func (user) InsertData() error { return nil }
