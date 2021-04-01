// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package dbinstall

import (
	"fmt"

	"github.com/palchukovsky/ss"
	ddbinstall "github.com/palchukovsky/ss/ddb/install"
)

////////////////////////////////////////////////////////////////////////////////

// ForEachTable calls the callback for each database table and logs it.
func ForEachTable(
	installer Installer,
	db ddbinstall.DB,
	callback func(ddbinstall.Table) error,
	log ss.ServiceLog,
) error {
	log.Debug("Debug prcessing each table...")

	tables := append(
		installer.NewTables(db),
		newConnectionTable(db),
		newUserTable(db))

	for _, table := range tables {
		log.Debug("Processing table %q...", table.GetName())
		if err := callback(table); err != nil {
			return fmt.Errorf(`failed to process %q: "%w"`, table.GetName(), err)
		}
		log.Info("Processing table %q completed.", table.GetName())
	}

	log.Debug("Processing each table successfully completed.")
	return nil
}

////////////////////////////////////////////////////////////////////////////////

// Installer describes the database installing interface.
type Installer interface {
	NewTables(ddbinstall.DB) []ddbinstall.Table
}

////////////////////////////////////////////////////////////////////////////////
