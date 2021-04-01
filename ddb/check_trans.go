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

// CheckTrans describes the part of WriteTrans witch builds check-expression.
type CheckTrans interface {
	WriteTransExpression
	Values(Values) CheckTrans
	Condition(string) CheckTrans
}

////////////////////////////////////////////////////////////////////////////////

func (trans *writeTrans) Check(key KeyRecord) CheckTrans {
	input := &dynamodb.ConditionCheck{
		TableName: aws.String(ss.S.NewBuildEntityName(key.GetTable())),
	}
	result := checkTrans{
		writeTransExpression: newWriteTransExpression(
			trans, dynamodb.TransactWriteItem{ConditionCheck: input}),
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
	result.input.ConditionExpression = aws.String(fmt.Sprintf(
		"attribute_exists(%s)", aliasReservedWord(
			key.GetKeyPartitionField(), &result.input.ExpressionAttributeNames)))
	return result
}

////////////////////////////////////////////////////////////////////////////////

type checkTrans struct {
	writeTransExpression
	input *dynamodb.ConditionCheck
}

func (trans checkTrans) Values(values Values) CheckTrans {
	trans.input.ExpressionAttributeValues = trans.marshalValues(values)
	return trans
}

func (trans checkTrans) Condition(condition string) CheckTrans {
	*trans.input.ConditionExpression += " and " +
		*aliasReservedInString(condition, &trans.input.ExpressionAttributeNames)
	return trans
}

////////////////////////////////////////////////////////////////////////////////
