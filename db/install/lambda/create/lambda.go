// Copyright 2021-2022, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package createdatabaselambda

import (
	"github.com/palchukovsky/ss"
	dbinstall "github.com/palchukovsky/ss/db/install"
	ddbinstall "github.com/palchukovsky/ss/ddb/install"
)

func Init(initService func(projectPackage string, params ss.ServiceParams)) {
	initService("install", ss.ServiceParams{})
}

func Run(installer dbinstall.Installer) {
	log := ss.S.Log()
	defer func() { log.CheckExit(recover()) }()
	log.Started()

	db := ddbinstall.NewDB()

	err := dbinstall.ForEachTable(
		installer,
		db,
		func(table ddbinstall.Table) error { return table.Create() },
		log)
	if err != nil {
		log.Panic(ss.NewLogMsg(`failed to create tables`).AddErr(err))
	}

	log.Info(ss.NewLogMsg("setupping..."))
	err = dbinstall.ForEachTable(
		installer,
		db,
		func(table ddbinstall.Table) error { return table.Setup() },
		log)
	if err != nil {
		log.Panic(ss.NewLogMsg(`failed to setup tables`).AddErr(err))
	}

	log.Info(ss.NewLogMsg("waiting..."))
	err = dbinstall.ForEachTable(
		installer,
		db,
		func(table ddbinstall.Table) error { return table.Wait() },
		log)
	if err != nil {
		log.Panic(ss.NewLogMsg(`failed to wait table`).AddErr(err))
	}

	log.Info(ss.NewLogMsg("inserting data..."))
	err = dbinstall.ForEachTable(
		installer,
		db,
		func(table ddbinstall.Table) error { return table.InsertData() },
		log)
	if err != nil {
		log.Panic(ss.NewLogMsg(`failed to insert data into table`).AddErr(err))
	}
}
