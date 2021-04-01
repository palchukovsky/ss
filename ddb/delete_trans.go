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
	input := &dynamodb.Delete{
		TableName: aws.String(ss.S.NewBuildEntityName(key.GetTable())),
	}
	result := deleteTrans{
		writeTransExpression: newWriteTransExpression(
			trans, dynamodb.TransactWriteItem{Delete: input}),
		input: input,
	}
	if trans.err != nil {
		return result
	}
	result.input.Key, trans.err = dynamodbattribute.MarshalMap(key.GetKey())
	if trans.err != nil {
		trans.err = fmt.Errorf(
			`failed to serialize key to delete from table %q: "%w", key: "%v"`,
			key.GetTable(), trans.err, key.GetKey())
	}
	result.input.ConditionExpression = aws.String(fmt.Sprintf(
		"attribute_exists(%s)", aliasReservedWord(
			key.GetKeyPartitionField(), &result.input.ExpressionAttributeNames)))
	return result
}

////////////////////////////////////////////////////////////////////////////////

type deleteTrans struct {
	writeTransExpression
	input *dynamodb.Delete
}

func (trans deleteTrans) Values(values Values) UpdateTrans {
	trans.input.ExpressionAttributeValues = trans.marshalValues(values)
	return trans
}

func (trans deleteTrans) Condition(condition string) UpdateTrans {
	*trans.input.ConditionExpression += " and " +
		*aliasReservedInString(condition, &trans.input.ExpressionAttributeNames)
	return trans
}

////////////////////////////////////////////////////////////////////////////////
