// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package dbinstall

import (
	"fmt"

	"github.com/palchukovsky/ss"
	ddbinstall "github.com/palchukovsky/ss/ddb/install"
)

// ForEachTable calls the callback for each database table and logs it.
func ForEachTable(
	installer Installer,
	db ddbinstall.DB,
	callback func(ddbinstall.Table) error,
	log ss.ServiceLog,
) error {
	log.Debug("Processing each table...")

	tables := append(
		installer.NewTables(db, log),
		newConnectionTable(db, log),
		newUserTable(db, log))
	defer func() {
		for _, table := range tables {
			table.Log().CheckExit()
		}
	}()

	for _, table := range tables {
		table.Log().Debug("Processing...")
		if err := callback(table); err != nil {
			return fmt.Errorf(`failed to process %q: "%w"`, table.GetName(), err)
		}
		table.Log().Info("Processed.")
	}

	log.Debug("Processing each table successfully completed.")
	return nil
}