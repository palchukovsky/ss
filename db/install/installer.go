// Copyright 2021-2022, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package dbinstall

import (
	"github.com/palchukovsky/ss"
	ddbinstall "github.com/palchukovsky/ss/ddb/install"
)

// Installer describes the database installing interface.
type Installer interface {
	NewTables(ddbinstall.DB, ss.Log) []ddbinstall.Table
	NewUserTableStreams() *ddbinstall.Streams
}
