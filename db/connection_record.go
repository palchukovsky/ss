// Copyright 2021-2022, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package db

import (
	"time"

	"github.com/palchukovsky/ss"
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

type ConnectionKeyValue struct {
	ID ss.ConnectionID `json:"id"`
}

func newConnectionKeyValue(id ss.ConnectionID) ConnectionKeyValue {
	return ConnectionKeyValue{ID: id}
}

type connectionKey struct {
	connectionRecord
	ConnectionKeyValue
}

func NewConnectionKey(id ss.ConnectionID) connectionKey {
	return connectionKey{ConnectionKeyValue: newConnectionKeyValue(id)}
}

func (key connectionKey) GetKey() interface{} { return key.ConnectionKeyValue }

////////////////////////////////////////////////////////////////////////////////

// Connection describes the record of a table with active connections.
type Connection struct {
	connectionRecord
	ConnectionKeyValue
	User           ss.UserID `json:"user"`
	Version        string    `json:"ver"`
	ExpirationTime ss.Time   `json:"expiration"`
}

// NewConnection creates new connection record.
func NewConnection(
	id ss.ConnectionID,
	user ss.UserID,
	version string,
) Connection {
	return Connection{
		ConnectionKeyValue: newConnectionKeyValue(id),
		User:               user,
		Version:            version,
		ExpirationTime:     ss.Now().Add(((time.Hour * 24) * 30) * 1),
	}
}

// GetData returns record's data.
func (record Connection) GetData() interface{} { return record }

////////////////////////////////////////////////////////////////////////////////
