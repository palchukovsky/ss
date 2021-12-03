// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ddb

import (
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/palchukovsky/ss"
)

// FindMany describes the interface to find many records records by key.
type FindMany interface {
	ss.NoCopy

	SetTable(RecordBuffer, []Key) CacheIterator
	Request()
}

////////////////////////////////////////////////////////////////////////////////

func (client *client) FindMany() FindMany {
	return &findMany{
		db: client.db,
		input: dynamodb.BatchGetItemInput{
			RequestItems: map[string]*dynamodb.KeysAndAttributes{},
		},
		output: map[string]*cacheIterator{},
	}
}

type findMany struct {
	ss.NoCopyImpl

	db     *dynamodb.DynamoDB
	input  dynamodb.BatchGetItemInput
	output map[string]*cacheIterator
}

func (find *findMany) SetTable(record RecordBuffer, keys []Key) CacheIterator {

	request := dynamodb.KeysAndAttributes{}

	request.ProjectionExpression = getRecordProjection(
		record,
		&request.ExpressionAttributeNames)

	for _, keySource := range keys {
		key, err := dynamodbattribute.MarshalMap(keySource)
		if err != nil {
			ss.S.Log().Panic(
				ss.
					NewLogMsg(`failed to serialize key`).
					AddErr(err).
					AddDump(record).
					AddDump(key))
		}
		request.Keys = append(request.Keys, key)
	}

	table := ss.S.NewBuildEntityName(record.GetTable())
	find.input.RequestItems[table] = &request
	result := newCacheIterator(nil, record)
	find.output[table] = result

	return result
}

func (find *findMany) Request() {

	request, response := find.db.BatchGetItemRequest(&find.input)
	if err := request.Send(); err != nil {
		ss.S.Log().Panic(
			ss.
				NewLogMsg(`failed to execute batch get item request`).
				AddErr(err).
				AddDump(find.input))
	}

	if len(response.UnprocessedKeys) != 0 {
		// In the 1st implementation UnprocessedKeys handling is not supported
		// to speed up development.
		ss.S.Log().Panic(
			ss.
				NewLogMsg(`batch get item request respons has unprocessed keys`).
				AddDump(response).
				AddDump(find.input))
	}

	for tableName, table := range response.Responses {
		find.output[tableName].Set(table)
	}
}
