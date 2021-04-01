// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ddbinstall

import (
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/palchukovsky/ss"
)

// DB describes database interface for installer.
type DB interface {
	CreateTable(dynamodb.CreateTableInput) error
	DescribeTable(dynamodb.DescribeTableInput,
	) (dynamodb.DescribeTableOutput, error)
	UpdateTable(dynamodb.UpdateTableInput) error
	UpdateTimeToLive(dynamodb.UpdateTimeToLiveInput) error
	DeleteTable(dynamodb.DeleteTableInput) error
	WaitTable(dynamodb.DescribeTableInput) error
	WaitUntilTableNotExists(dynamodb.DescribeTableInput) error
}

////////////////////////////////////////////////////////////////////////////////

func NewDB() DB { return dbClient{db: dynamodb.New(ss.S.GetAWSSessionV1())} }

type dbClient struct{ db *dynamodb.DynamoDB }

func (db dbClient) CreateTable(input dynamodb.CreateTableInput) error {
	request, _ := db.db.CreateTableRequest(&input)
	return request.Send()
}

func (db dbClient) DescribeTable(input dynamodb.DescribeTableInput,
) (dynamodb.DescribeTableOutput, error) {
	request, result := db.db.DescribeTableRequest(&input)
	if err := request.Send(); err != nil {
		return dynamodb.DescribeTableOutput{}, err
	}
	return *result, nil
}

func (db dbClient) WaitTable(input dynamodb.DescribeTableInput) error {
	return db.db.WaitUntilTableExists(&input)
}

func (db dbClient) WaitUntilTableNotExists(input dynamodb.DescribeTableInput,
) error {
	return db.db.WaitUntilTableNotExists(&input)
}

func (db dbClient) UpdateTable(input dynamodb.UpdateTableInput) error {
	request, _ := db.db.UpdateTableRequest(&input)
	return request.Send()
}

func (db dbClient) UpdateTimeToLive(input dynamodb.UpdateTimeToLiveInput) error {
	request, _ := db.db.UpdateTimeToLiveRequest(&input)
	return request.Send()
}

func (db dbClient) DeleteTable(input dynamodb.DeleteTableInput) error {
	request, _ := db.db.DeleteTableRequest(&input)
	return request.Send()
}
