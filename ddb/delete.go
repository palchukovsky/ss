// Copyright 2021-2022, the SS project owners. All rights reserved.
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
	CheckedExpression

	Condition(string) Delete
	Values(Values) Delete

	RequestWithResult() Result
	RequestAndReturn(RecordBuffer) Result
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
	checkedExpression

	db    *dynamodb.DynamoDB
	input dynamodb.DeleteItemInput
}

func (client *client) newDeleteTrans(key KeyRecord) *delete {
	result := delete{
		checkedExpression: newCheckedExpression(),
		db:                client.db,
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

func (trans *delete) RequestWithResult() Result {
	result, _ := trans.request()
	return result
}

func (trans *delete) RequestAndReturn(resultRecord RecordBuffer) Result {
	trans.input.ReturnValues = aws.String(dynamodb.ReturnValueAllOld)
	result, output := trans.request()
	if !result.IsSuccess() {
		return result
	}
	err := dynamodbattribute.UnmarshalMap(output.Attributes, resultRecord)
	if err != nil {
		ss.S.Log().Panic(
			ss.
				NewLogMsg(
					`failed to read delete response from table %q`,
					trans.getTable()).
				AddErr(err).
				AddDump(result))
	}
	return result
}

func (trans *delete) request() (Result, *dynamodb.DeleteItemOutput) {
	request, output := trans.db.DeleteItemRequest(&trans.input)
	result, err := newResult(request.Send(), trans.isConditionalCheckFailAllowed)
	if err != nil {
		ss.S.Log().Panic(
			ss.
				NewLogMsg(`failed to delete item from table %q`, trans.getTable()).
				AddErr(err).
				AddDump(trans.input))
	}
	return result, output
}

func (delete *delete) getTable() string { return *delete.input.TableName }
