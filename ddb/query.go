// Copyright 2021-2022, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ddb

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/palchukovsky/ss"
)

// Query describes the interface to query records from a database.
type Query interface {
	ss.NoCopy

	Filter(string) Query
	Limit(int64) Query
	Descending() Query
	RequestOne() bool
	RequestPaged() Iterator
	RequestAll() CacheIterator
}

////////////////////////////////////////////////////////////////////////////////

func (client *client) Query(
	record RecordBuffer,
	keyCondition string,
	values Values,
) Query {
	return newQuery(client, record, keyCondition, values)
}

func newQuery(
	client *client,
	record RecordBuffer,
	keyCondition string,
	values Values,
) *query {
	result := &query{
		client: client,
		Record: record,
		Input: dynamodb.QueryInput{
			TableName: aws.String(ss.S.NewBuildEntityName(record.GetTable())),
		},
	}
	values.Marshal(&result.Input.ExpressionAttributeValues)

	result.Input.KeyConditionExpression = aliasReservedInString(
		keyCondition,
		&result.Input.ExpressionAttributeNames)
	result.Input.ProjectionExpression = getRecordProjection(
		record,
		&result.Input.ExpressionAttributeNames)

	return result
}

type query struct {
	ss.NoCopyImpl

	client *client             `json:"-"`
	Record RecordBuffer        `json:"record"`
	Input  dynamodb.QueryInput `json:"input"`
}

func (query *query) Filter(filter string) Query {
	query.Input.FilterExpression = aliasReservedInString(
		filter,
		&query.Input.ExpressionAttributeNames)
	return query
}

func (query *query) Limit(limit int64) Query {
	query.Input.Limit = &limit
	return query
}

func (query *query) Descending() Query {
	query.Input.ScanIndexForward = ss.BoolPtr(false)
	return query
}

func (query *query) RequestPaged() Iterator {
	request, _ := query.client.db.QueryRequest(&query.Input)
	return newPagedIterator(request, query.Record)
}

func (query *query) RequestOne() bool {
	it := query.RequestAll()
	if it.GetSize() > 1 {
		ss.S.Log().Panic(
			ss.NewLogMsg(`expected one record, but returned %d`, it.GetSize()))
	}
	return it.Next()
}

func (query *query) RequestAll() CacheIterator {
	request, output := query.client.db.QueryRequest(&query.Input)
	if err := request.Send(); err != nil {
		ss.S.Log().Panic(
			ss.
				NewLogMsg(`failed to query from table %q`, query.Record.GetTable()).
				AddErr(err).
				AddDump(query.Input))
	}
	return newCacheIterator(output.Items, query.Record)
}
