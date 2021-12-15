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

	WriteWithResult(WriteTrans) TransResult
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

func (client *client) WriteWithResult(trans WriteTrans) TransResult {
	request, _ := client.db.TransactWriteItemsRequest(trans.Result())
	result, err := newTransResult(request.Send())
	if err != nil {
		ss.S.Log().Panic(
			ss.
				NewLogMsg("failed to write DDB transaction").
				AddDump(trans).
				AddErr(err))
	}
	return result
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
