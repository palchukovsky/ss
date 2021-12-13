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

// Delete describes the interface to delete one record by key.
type Delete interface {
	ss.NoCopy

	Condition(string) Delete
	Values(Values) Delete

	// Request executed request and return false if failed to find a record
	// or if added conditions are failed.
	Request() bool
	// RequestAndReturn executed request and return false if failed to find
	// a record or if added conditions are failed.
	RequestAndReturn(RecordBuffer) bool
}

////////////////////////////////////////////////////////////////////////////////

func (client *client) Delete(key KeyRecord) Delete {
	result := client.newDeleteTrans(key)
	result.input.ConditionExpression = aws.String(
		fmt.Sprintf(
			"attribute_exists(%s)",
			aliasReservedWord(
				key.GetKeyPartitionField(),
				&result.input.ExpressionAttributeNames)))
	return result
}

func (client *client) DeleteIfExisting(key KeyRecord) Delete {
	return client.newDeleteTrans(key)
}

type delete struct {
	ss.NoCopyImpl

	db    *dynamodb.DynamoDB
	input dynamodb.DeleteItemInput
}

func (client *client) newDeleteTrans(key KeyRecord) *delete {
	result := delete{
		db: client.db,
		input: dynamodb.DeleteItemInput{
			TableName: aws.String(ss.S.NewBuildEntityName(key.GetTable())),
		},
	}
	var err error
	result.input.Key, err = dynamodbattribute.MarshalMap(key.GetKey())
	if err != nil {
		ss.S.Log().Panic(
			ss.
				NewLogMsg(
					`failed to serialize key to delete from table %q`,
					result.getTable()).
				AddErr(err).
				AddDump(key))
	}
	return &result
}

func (trans *delete) Values(values Values) Delete {
	values.Marshal(&trans.input.ExpressionAttributeValues)
	return trans
}

func (trans *delete) Condition(condition string) Delete {
	*trans.input.ConditionExpression += " and (" +
		*aliasReservedInString(condition, &trans.input.ExpressionAttributeNames) +
		")"
	return trans
}

func (trans *delete) Request() bool {
	return trans.request() != nil
}

func (trans *delete) RequestAndReturn(result RecordBuffer) bool {
	trans.input.ReturnValues = aws.String(dynamodb.ReturnValueAllOld)
	output := trans.request()
	if output == nil {
		return false
	}
	err := dynamodbattribute.UnmarshalMap(output.Attributes, result)
	if err != nil {
		ss.S.Log().Panic(
			ss.
				NewLogMsg(
					`failed to read delete response from table %q`,
					trans.getTable()).
				AddErr(err).
				AddDump(result))
	}
	return true
}

func (trans *delete) request() *dynamodb.DeleteItemOutput {
	request, result := trans.db.DeleteItemRequest(&trans.input)
	if err := request.Send(); err != nil {
		if trans.input.ConditionExpression != nil {
			if isConditionalCheckError(err) {
				// no error, but not found
				return nil
			}
		}
		ss.S.Log().Panic(
			ss.
				NewLogMsg(`failed to delete item from table %q`, trans.getTable()).
				AddErr(err).
				AddDump(trans.input))
	}
	return result
}

func (delete *delete) getTable() string { return *delete.input.TableName }
