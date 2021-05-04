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
	log ss.Log,
) error {
	log.Debug(ss.NewLogMsg("processing each table..."))

	tables := append(
		installer.NewTables(db, log),
		newConnectionTable(db, log),
		newUserTable(db, log))

	for _, table := range tables {
		table.Log().Debug(ss.NewLogMsg("processing..."))
		if err := callback(table); err != nil {
			return fmt.Errorf(`failed to process %q: "%w"`, table.GetName(), err)
		}
		table.Log().Info(ss.NewLogMsg("processed"))
	}

	log.Debug(ss.NewLogMsg("processing each table successfully completed"))
	return nil
}
