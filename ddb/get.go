// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ddb

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/palchukovsky/ss"
)

// Find describes the interface to find one record by key.
type Find interface{ Request() (bool, error) }

// Get describes the interface to query one record by key from a database.
type Get interface{ Request() error }

////////////////////////////////////////////////////////////////////////////////

func newFind(db *dynamodb.DynamoDB, record KeyRecordBuffer) find {
	result := find{
		db:     db,
		record: record,
		input: dynamodb.GetItemInput{
			TableName: aws.String(ss.S.NewBuildEntityName(record.GetTable())),
		},
	}
	result.input.Key, result.err = dynamodbattribute.MarshalMap(record.GetKey())
	if result.err != nil {
		result.err = fmt.Errorf(
			`failed to serialize key to get item from %q: "%w", key: "%v"`,
			record.GetTable(), result.err, record.GetKey())
		return result
	}
	result.input.ProjectionExpression = getRecordProjection(record,
		&result.input.ExpressionAttributeNames)
	return result
}

type find struct {
	db     *dynamodb.DynamoDB
	record RecordBuffer
	input  dynamodb.GetItemInput
	err    error
}

func (find find) Request() (bool, error) {
	if find.err != nil {
		return false, find.err
	}
	request, response := find.db.GetItemRequest(&find.input)
	if err := request.Send(); err != nil {
		return false, fmt.Errorf(
			`failed to get item from table %q: "%w", input: %v`,
			find.record.GetTable(), err, ss.Dump(find.input))
	}
	if len(response.Item) == 0 {
		return false, nil
	}
	err := dynamodbattribute.UnmarshalMap(response.Item, find.record)
	if err != nil {
		return false, fmt.Errorf(
			`failed to read get-response from table %q: "%w", dump: %s`,
			find.record.GetTable(), err, ss.Dump(response.Item))
	}
	return true, nil
}

////////////////////////////////////////////////////////////////////////////////

func newGet(db *dynamodb.DynamoDB, record KeyRecordBuffer) get {
	return get{find: newFind(db, record)}
}

type get struct{ find }

func (get get) Request() error {
	isFound, err := get.find.Request()
	if err != nil {
		return err
	}
	if !isFound {
		return fmt.Errorf(`unknown record in table %q by input %s`,
			get.find.record.GetTable(), ss.Dump(get.find.input))
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////
