// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package db

import (
	"github.com/palchukovsky/ss"
	"github.com/palchukovsky/ss/ddb"
)

////////////////////////////////////////////////////////////////////////////////

type ConnectionIDByUser struct {
	ConnectionUserIndex
	ID ss.ConnectionID `json:"id"`
}

func (record *ConnectionIDByUser) Clear() { *record = ConnectionIDByUser{} }

// IteratorMover describes intreface to move iterator.
type FindUserConnectionsIterator interface {
	Next() bool
	Get() ss.ConnectionID
}

func FindUserConnections(
	user ss.UserID,
	db ddb.Client,
) FindUserConnectionsIterator {
	var result findUserConnectionsIterator
	result.it = db.
		Index(&result.buffer).
		Query("user = :u", ddb.Values{":u": user}).
		RequestPaged()
	return &result
}

type findUserConnectionsIterator struct {
	it     ddb.Iterator
	buffer ConnectionIDByUser
}

func (it findUserConnectionsIterator) Next() bool { return it.it.Next() }

func (it findUserConnectionsIterator) Get() ss.ConnectionID {
	return it.buffer.ID
}

////////////////////////////////////////////////////////////////////////////////
