// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package db

import (
	"time"

	"github.com/palchukovsky/ss"
	"github.com/palchukovsky/ss/ddb"
)

////////////////////////////////////////////////////////////////////////////////

type connectionRecord struct{}

// GetTable returns table name.
func (connectionRecord) GetTable() string { return "Connection" }

// GetKeyPartitionField returns partition field name.
func (connectionRecord) GetKeyPartitionField() string { return "id" }

// GetKeySortField returns sort field name.
func (connectionRecord) GetKeySortField() string { return "" }

////////////////////////////////////////////////////////////////////////////////

// ConnectionUserIndex describes connection index "bu user".
type ConnectionUserIndex struct{ connectionRecord }

// GetIndex returns index name.
func (ConnectionUserIndex) GetIndex() string { return "User" }

// GetIndexPartitionField returns index partition field name.
func (ConnectionUserIndex) GetIndexPartitionField() string { return "user" }

// GetIndexSortField returns index sort field name.
func (ConnectionUserIndex) GetIndexSortField() string { return "" }

func (ConnectionUserIndex) GetProjection() []string { return []string{} }

////////////////////////////////////////////////////////////////////////////////

type connectionKey struct {
	connectionRecord
	id ss.ConnectionID
}

func NewConnectionKey(id ss.ConnectionID) connectionKey {
	return connectionKey{id: id}
}

func (key connectionKey) GetKey() interface{} {
	return struct {
		ID ss.ConnectionID `json:"id"`
	}{ID: key.id}
}

////////////////////////////////////////////////////////////////////////////////

// Connection describes the record of a table with active connections.
type Connection struct {
	connectionRecord
	ID             ss.ConnectionID `json:"id"`
	User           ss.UserID       `json:"user"`
	ExpirationTime ddb.Time        `json:"expired"`
}

// NewConnection creates new connection record.
func NewConnection(id ss.ConnectionID, user ss.UserID) Connection {
	return Connection{
		ID:             id,
		User:           user,
		ExpirationTime: ddb.Now().Add(((time.Hour * 24) * 30) * 1),
	}
}

// GetData returns record's data.
func (record Connection) GetData() interface{} { return record }

////////////////////////////////////////////////////////////////////////////////
