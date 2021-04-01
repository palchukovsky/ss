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

func Init(serviceInit func(projectPackage string)) { serviceInit("install") }

func Run(installer dbinstall.Installer) {
	log := ss.S.Log().NewSession("database.delete")
	defer log.CheckExit()
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
				log.Info("Table %q doesn't exist: %q.", table.GetName(), err)
			}
			return nil
		},
		log)
	if err != nil {
		log.Panic(`Failed to delete database: "%v".`, err)
	}

	log.Info("Waition for deletion...")
	err = dbinstall.ForEachTable(
		installer,
		db,
		func(table ddbinstall.Table) error { return table.WaitUntilNotExists() },
		log)
	if err != nil {
		log.Panic(`Failed to wait table deletion: "%v".`, err)
	}
}
