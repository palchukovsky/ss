// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package db

import (
	"github.com/palchukovsky/ss"
	"github.com/palchukovsky/ss/ddb"
	"github.com/palchukovsky/ss/lambda"
)

////////////////////////////////////////////////////////////////////////////////

type ConnectionIDByUser struct {
	ConnectionUserIndex
	ID lambda.ConnectionID `json:"id"`
}

func (record *ConnectionIDByUser) Clear() { *record = ConnectionIDByUser{} }

// IteratorMover describes intreface to move iterator.
type FindUserConnectionsIterator interface {
	Next() bool
	Get() lambda.ConnectionID
	Err() error
}

func FindUserConnections(
	user ss.UserID,
	db ddb.Client,
) (
	FindUserConnectionsIterator,
	error,
) {
	var result findUserConnectionsIterator
	var err error
	result.it, err = db.
		Index(&result.buffer).
		Query("user = :u", ddb.Values{":u": user}).
		RequestPaged()
	return result, err
}

type findUserConnectionsIterator struct {
	it     ddb.Iterator
	buffer ConnectionIDByUser
}

func (it findUserConnectionsIterator) Next() bool {
	return it.it.Next()
}
func (it findUserConnectionsIterator) Get() lambda.ConnectionID {
	return it.buffer.ID
}
func (it findUserConnectionsIterator) Err() error {
	return it.it.Err()
}

////////////////////////////////////////////////////////////////////////////////
