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

// UpdateTrans describes the part of WriteTrans witch builds update-expression.
type UpdateTrans interface {
	WriteTransExpression
	Values(Values) UpdateTrans
	Condition(string) UpdateTrans
}

////////////////////////////////////////////////////////////////////////////////

func (trans *writeTrans) Update(key KeyRecord, update string) UpdateTrans {
	input := &dynamodb.Update{
		TableName: aws.String(ss.S.NewBuildEntityName(key.GetTable()))}
	result := updateTrans{
		writeTransExpression: newWriteTransExpression(
			trans, dynamodb.TransactWriteItem{Update: input}),
		input: input,
	}
	if trans.err != nil {
		return result
	}
	result.input.Key, trans.err = dynamodbattribute.MarshalMap(key.GetKey())
	if trans.err != nil {
		trans.err = fmt.Errorf(
			`failed to serialize key to update table %q: "%w", key: "%v"`,
			key.GetTable(), trans.err, key.GetKey())
	}
	result.input.UpdateExpression = aliasReservedInString(update,
		&result.input.ExpressionAttributeNames)
	result.input.ConditionExpression = aws.String(fmt.Sprintf(
		"attribute_exists(%s)", aliasReservedWord(
			key.GetKeyPartitionField(), &result.input.ExpressionAttributeNames)))
	return result
}

////////////////////////////////////////////////////////////////////////////////

type updateTrans struct {
	writeTransExpression
	input *dynamodb.Update
}

func (trans updateTrans) Values(values Values) UpdateTrans {
	trans.input.ExpressionAttributeValues = trans.marshalValues(values)
	return trans
}

func (trans updateTrans) Condition(condition string) UpdateTrans {
	*trans.input.ConditionExpression += " and " +
		*aliasReservedInString(condition, &trans.input.ExpressionAttributeNames)
	return trans
}

////////////////////////////////////////////////////////////////////////////////
