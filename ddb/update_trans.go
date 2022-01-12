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

// UpdateTrans describes the part of WriteTrans witch builds update-expression.
type UpdateTrans interface {
	WriteTransExpression
	Values(Values) UpdateTrans
	Value(name string, value interface{}) UpdateTrans
	Alias(name, value string) UpdateTrans
	Condition(string) UpdateTrans
}

////////////////////////////////////////////////////////////////////////////////

func (trans *writeTrans) Update(key KeyRecord, update string) UpdateTrans {
	input := &dynamodb.Update{
		TableName: aws.String(ss.S.NewBuildEntityName(key.GetTable())),
	}
	result := updateTrans{
		writeTransExpression: newWriteTransExpression(
			trans,
			dynamodb.TransactWriteItem{Update: input}),
		input: input,
	}
	var err error
	result.input.Key, err = dynamodbattribute.MarshalMap(key.GetKey())
	if err != nil {
		ss.S.Log().Panic(
			ss.
				NewLogMsg(
					"failed to serialize key to update table %q",
					key.GetTable()).
				AddDump(key).
				AddDump(input).
				AddErr(err))
	}
	result.input.UpdateExpression = aliasReservedInString(
		update,
		&result.input.ExpressionAttributeNames)
	result.input.ConditionExpression = aws.String(
		fmt.Sprintf(
			"attribute_exists(%s)",
			aliasReservedWord(
				key.GetKeyPartitionField(),
				&result.input.ExpressionAttributeNames)))
	return &result
}

////////////////////////////////////////////////////////////////////////////////

type updateTrans struct {
	writeTransExpression
	input *dynamodb.Update
}

func (trans *updateTrans) Values(values Values) UpdateTrans {
	trans.marshalValues(values, &trans.input.ExpressionAttributeValues)
	return trans
}

func (trans *updateTrans) Value(name string, value interface{}) UpdateTrans {
	return trans.Values(Values{name: value})
}

func (trans *updateTrans) Alias(name, value string) UpdateTrans {
	trans.addAlias(name, value, &trans.input.ExpressionAttributeNames)
	return trans
}

func (trans *updateTrans) Condition(condition string) UpdateTrans {
	*trans.input.ConditionExpression += " and (" +
		*aliasReservedInString(condition, &trans.input.ExpressionAttributeNames) +
		")"
	return trans
}

////////////////////////////////////////////////////////////////////////////////
