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
	hasUserUpdateLambda bool
}

func newUserTable(
	ddb ddbinstall.DB,
	log ss.Log,
	hasUserUpdateLambda bool,
) ddbinstall.Table {
	return user{
		TableAbstraction:    ddbinstall.NewTableAbstraction(ddb, db.User{}, log),
		hasUserUpdateLambda: hasUserUpdateLambda,
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

	if table.hasUserUpdateLambda {
		err := table.EnableStreams(
			ddbinstall.NewStreams(
				// Full view required to separate business logic record inserts
				// and deletes from index records inserts and deletes:
				ddbinstall.StreamViewTypeFull,
				ddbinstall.NewStream("UserContentUpdate"),
			),
		)
		if err != nil {
			return err
		}
	}

	return nil
}

func (user) InsertData() error { return nil }
