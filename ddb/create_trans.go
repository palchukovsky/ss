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

// CreateTrans describes the part of WriteTrans witch builds create-expression.
type CreateTrans interface {
	WriteTransExpression
	Values(Values) CreateTrans
	Value(name string, value interface{}) CreateTrans
	Alias(name, value string) CreateTrans
	Condition(string) CreateTrans
}

////////////////////////////////////////////////////////////////////////////////

func (trans *writeTrans) CreateIfNotExists(record DataRecord) CreateTrans {
	result := trans.newCreateTrans(record)
	result.Condition(
		fmt.Sprintf("attribute_not_exists(%s)", record.GetKeyPartitionField()))
	return result
}

func (trans *writeTrans) CreateOrReplace(record DataRecord) CreateTrans {
	return trans.newCreateTrans(record)
}

func (trans *writeTrans) Replace(record DataRecord) CreateTrans {
	result := trans.newCreateTrans(record)
	result.Condition(
		fmt.Sprintf("attribute_exists(%s)", record.GetKeyPartitionField()))
	return result
}

////////////////////////////////////////////////////////////////////////////////

type createTrans struct {
	writeTransExpression
	input *dynamodb.Put
}

func (trans *writeTrans) newCreateTrans(record DataRecord) *createTrans {
	input := &dynamodb.Put{
		TableName: aws.String(ss.S.NewBuildEntityName(record.GetTable())),
	}
	result := createTrans{
		writeTransExpression: newWriteTransExpression(
			trans,
			dynamodb.TransactWriteItem{Put: input}),
		input: input,
	}
	var err error
	result.input.Item, err = dynamodbattribute.MarshalMap(record.GetData())
	if err != nil {
		ss.S.Log().Panic(
			ss.
				NewLogMsg(
					"failed to serialize item to put into table %q",
					record.GetTable()).
				AddDump(record).
				AddDump(input).
				AddErr(err))
	}
	return &result
}

func (trans *createTrans) Values(values Values) CreateTrans {
	trans.marshalValues(values, &trans.input.ExpressionAttributeValues)
	return trans
}

func (trans *createTrans) Value(name string, value interface{}) CreateTrans {
	return trans.Values(Values{name: value})
}

func (trans *createTrans) Alias(name, value string) CreateTrans {
	trans.addAlias(name, value, &trans.input.ExpressionAttributeNames)
	return trans
}

func (trans *createTrans) Condition(condition string) CreateTrans {
	condition = "(" +
		*aliasReservedInString(condition, &trans.input.ExpressionAttributeNames) +
		")"
	if trans.input.ConditionExpression == nil {
		trans.input.ConditionExpression = &condition
	} else {
		*trans.input.ConditionExpression += " and " + condition
	}
	return trans
}

////////////////////////////////////////////////////////////////////////////////
