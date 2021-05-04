// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package deletetabaselambda

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
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
		func(table ddbinstall.Table) error {
			if err := table.Delete(); err != nil {
				var awsErr awserr.Error
				if !errors.As(err, &awsErr) {
					return err
				}
				if awsErr.Code() != dynamodb.ErrCodeResourceNotFoundException {
					return err
				}
				table.Log().Info(ss.NewLogMsg("table doesn't exist").AddErr(err))
			}
			return nil
		},
		log)
	if err != nil {
		log.Panic(ss.NewLogMsg(`failed to delete database`).AddErr(err))
	}

	log.Info(ss.NewLogMsg("waiting for deletion..."))
	err = dbinstall.ForEachTable(
		installer,
		db,
		func(table ddbinstall.Table) error { return table.WaitUntilNotExists() },
		log)
	if err != nil {
		log.Panic(ss.NewLogMsg(`failed to wait table deletion`).AddErr(err))
	}
}
