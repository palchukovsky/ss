// Copyright 2021-2022, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ddb

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/palchukovsky/ss"
)

// Find describes the interface to find one record by key.
type Find interface {
	ss.NoCopy

	Request() bool
}

// Get describes the interface to query one record by key from a database.
type Get interface{ Request() }

////////////////////////////////////////////////////////////////////////////////

func newFind(db *dynamodb.DynamoDB, record KeyRecordBuffer) *find {
	result := find{
		db:     db,
		record: record,
		input: dynamodb.GetItemInput{
			TableName: aws.String(ss.S.NewBuildEntityName(record.GetTable())),
		},
	}
	var err error
	result.input.Key, err = dynamodbattribute.MarshalMap(record.GetKey())
	if err != nil {
		ss.S.Log().Panic(
			ss.
				NewLogMsg(
					`failed to serialize key to get item from %q`,
					record.GetTable()).
				AddErr(err).
				AddDump(record))
		return &result
	}
	result.input.ProjectionExpression = getRecordProjection(
		record,
		&result.input.ExpressionAttributeNames)
	return &result
}

type find struct {
	ss.NoCopyImpl

	db     *dynamodb.DynamoDB
	record RecordBuffer
	input  dynamodb.GetItemInput
}

func (find *find) Request() bool {
	request, response := find.db.GetItemRequest(&find.input)
	if err := request.Send(); err != nil {
		ss.S.Log().Panic(
			ss.
				NewLogMsg(`failed to get item from table %q`, find.record.GetTable()).
				AddErr(err).
				AddDump(find.record).
				AddDump(find.input))
	}
	if len(response.Item) == 0 {
		return false
	}
	err := dynamodbattribute.UnmarshalMap(response.Item, find.record)
	if err != nil {
		ss.S.Log().Panic(
			ss.
				NewLogMsg(
					`failed to read get-response from table %q`,
					find.record.GetTable()).
				AddErr(err).
				AddDump(find.record).
				AddDump(find.input))
	}
	return true
}

////////////////////////////////////////////////////////////////////////////////

func newGet(db *dynamodb.DynamoDB, record KeyRecordBuffer) *get {
	return &get{find: newFind(db, record)}
}

type get struct{ *find }

func (get *get) Request() {
	if isFound := get.find.Request(); !isFound {
		ss.S.Log().Panic(
			ss.
				NewLogMsg(`unknown record in table %q`, get.find.record.GetTable()).
				AddDump(get.record).
				AddDump(get.input))
	}
}

////////////////////////////////////////////////////////////////////////////////
