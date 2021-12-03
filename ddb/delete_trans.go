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

// DeleteTrans describes the part of WriteTrans witch builds delete-expression.
type DeleteTrans interface {
	WriteTransExpression
	Values(Values) UpdateTrans
	Condition(string) UpdateTrans
}

////////////////////////////////////////////////////////////////////////////////

func (trans *writeTrans) Delete(key KeyRecord) DeleteTrans {
	result := trans.newDeleteTrans(key)
	result.input.ConditionExpression = aws.String(
		fmt.Sprintf(
			"attribute_exists(%s)",
			aliasReservedWord(
				key.GetKeyPartitionField(),
				&result.input.ExpressionAttributeNames)))
	return result
}

func (trans *writeTrans) DeleteIfExisting(key KeyRecord) DeleteTrans {
	return trans.newDeleteTrans(key)
}

////////////////////////////////////////////////////////////////////////////////

type deleteTrans struct {
	writeTransExpression
	input *dynamodb.Delete
}

func (trans *writeTrans) newDeleteTrans(key KeyRecord) *deleteTrans {
	input := &dynamodb.Delete{
		TableName: aws.String(ss.S.NewBuildEntityName(key.GetTable())),
	}
	result := &deleteTrans{
		writeTransExpression: newWriteTransExpression(
			trans, dynamodb.TransactWriteItem{Delete: input}),
		input: input,
	}
	var err error
	result.input.Key, err = dynamodbattribute.MarshalMap(key.GetKey())
	if err != nil {
		ss.S.Log().Panic(
			ss.
				NewLogMsg(
					"failed to serialize key to delete from table %q",
					key.GetTable()).
				AddDump(key).
				AddDump(input).
				AddErr(err))
	}
	return result
}

func (trans *deleteTrans) Values(values Values) UpdateTrans {
	trans.marshalValues(values, &trans.input.ExpressionAttributeValues)
	return trans
}

func (trans *deleteTrans) Value(name string, value interface{}) UpdateTrans {
	return trans.Values(Values{name: value})
}

func (trans *deleteTrans) Alias(name, value string) UpdateTrans {
	trans.addAlias(name, value, &trans.input.ExpressionAttributeNames)
	return trans
}

func (trans *deleteTrans) Condition(condition string) UpdateTrans {
	*trans.input.ConditionExpression += " and (" +
		*aliasReservedInString(condition, &trans.input.ExpressionAttributeNames) +
		")"
	return trans
}

////////////////////////////////////////////////////////////////////////////////
