// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package createdatabaselambda

import (
	"github.com/palchukovsky/ss"
	dbinstall "github.com/palchukovsky/ss/db/install"
	ddbinstall "github.com/palchukovsky/ss/ddb/install"
)

func Init(serviceInit func(projectPackage string)) { serviceInit("install") }

func Run(installer dbinstall.Installer) {
	log := ss.S.Log()
	defer log.CheckExit()
	log.Started()

	db := ddbinstall.NewDB()

	err := dbinstall.ForEachTable(
		installer,
		db,
		func(table ddbinstall.Table) error { return table.Create() },
		log)
	if err != nil {
		log.Panic(`Failed to create tables: "%v".`, err)
	}

	log.Info("Setupping...")
	err = dbinstall.ForEachTable(
		installer,
		db,
		func(table ddbinstall.Table) error { return table.Setup() },
		log)
	if err != nil {
		log.Panic(`Failed to setup tables: "%v".`, err)
	}

	log.Info("Waiting...")
	err = dbinstall.ForEachTable(
		installer,
		db,
		func(table ddbinstall.Table) error { return table.Wait() },
		log)
	if err != nil {
		log.Panic(`Failed to wait table: "%v".`, err)
	}
}
