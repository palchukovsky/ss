// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ddb

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/palchukovsky/ss"
)

// Client is a database client interface.
type Client interface {
	ss.NoCopy

	Index(resultRecord IndexRecord) Index

	Get(KeyRecordBuffer) Get
	Find(KeyRecordBuffer) Find
	FindMany() FindMany
	Query(record RecordBuffer, keyCondition string, values Values) Query

	Create(data DataRecord) Create
	CreateIfNotExists(data DataRecord) CreateIfNotExists
	CreateOrReplace(data DataRecord) Create
	Update(key KeyRecord) Update
	Delete(key KeyRecord) Delete
	DeleteIfExisting(key KeyRecord) Delete

	Write(WriteTrans)
	WriteConditioned(trans WriteTrans) bool
	WriteConditionedWithResult(
		trans WriteTrans,
		conditionFromIndex int,
		conditionsNumber int,
	) []bool
}

// GetClientInstance returns reference to client singleton.
func GetClientInstance() Client {
	if clientInstance == nil {
		clientInstance = &client{db: dynamodb.New(ss.S.NewAWSSessionV1())}
	}
	return clientInstance
}

// Index describes db-command interface for the table index.
type Index interface {
	Query(keyCondition string, values Values) Query
}

////////////////////////////////////////////////////////////////////////////////

var clientInstance Client

type client struct {
	ss.NoCopyImpl

	db *dynamodb.DynamoDB
}

func (client *client) DynamoDB() *dynamodb.DynamoDB { return client.db }

func (client *client) Index(record IndexRecord) Index {
	return &index{client: client, record: record}
}

func (client *client) Find(key KeyRecordBuffer) Find {
	return newFind(client.db, key)
}
func (client *client) Get(key KeyRecordBuffer) Get {
	return newGet(client.db, key)
}

func (client *client) Write(trans WriteTrans) {
	request, _ := client.db.TransactWriteItemsRequest(trans.Result())
	if err := request.Send(); err != nil {
		ss.S.Log().Panic(
			ss.
				NewLogMsg("failed to write DDB transaction").
				AddDump(trans).
				AddErr(err))
	}
}

func (client *client) WriteConditioned(trans WriteTrans) bool {

	request, _ := client.db.TransactWriteItemsRequest(trans.Result())

	err := request.Send()
	if err == nil {
		return true
	}

	if !isConditionalCheckError(err) {
		ss.S.Log().Panic(
			ss.
				NewLogMsg("failed to write DDB transaction with conditions").
				AddDump(trans).
				AddErr(err))
		// never reaches
	}
	return false
}

func (client *client) WriteConditionedWithResult(
	trans WriteTrans,
	conditionFromIndex int,
	conditionsNumber int,
) []bool {

	request, _ := client.db.TransactWriteItemsRequest(trans.Result())

	err := request.Send()
	if err == nil {
		return nil
	}

	failedConditions := parseErrorConditionalCheckFailed(
		err,
		conditionFromIndex,
		conditionsNumber)
	if failedConditions != nil {
		return failedConditions
	}

	ss.S.Log().Panic(
		ss.
			NewLogMsg("failed to write DDB transaction with conditions").
			AddDump(trans).
			AddErr(err))
	return nil // never reaches
}

////////////////////////////////////////////////////////////////////////////////

type index struct {
	ss.NoCopyImpl

	client *client
	record IndexRecord
}

func (index *index) Query(keyCondition string, values Values) Query {
	result := newQuery(index.client, index.record, keyCondition, values)
	result.Input.IndexName = aws.String(index.record.GetIndex())
	return result
}

////////////////////////////////////////////////////////////////////////////////
