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
}

////////////////////////////////////////////////////////////////////////////////

func (trans *writeTrans) Create(record DataRecord) CreateTrans {
	result := trans.newCreateTrans(record)
	result.input.ConditionExpression = aws.String(fmt.Sprintf(
		"attribute_not_exists(%s)",
		aliasReservedWord(
			record.GetKeyPartitionField(),
			&result.input.ExpressionAttributeNames)))
	return result
}

func (trans *writeTrans) CreateOrUpdate(record DataRecord) CreateTrans {
	return trans.newCreateTrans(record)
}

////////////////////////////////////////////////////////////////////////////////

type createTrans struct {
	writeTransExpression
	input *dynamodb.Put
}

////////////////////////////////////////////////////////////////////////////////

func (trans *writeTrans) newCreateTrans(record DataRecord) createTrans {
	input := &dynamodb.Put{
		TableName: aws.String(ss.S.NewBuildEntityName(record.GetTable())),
	}
	result := createTrans{
		writeTransExpression: newWriteTransExpression(
			trans, dynamodb.TransactWriteItem{Put: input}),
		input: input,
	}
	result.input.Item, trans.err = dynamodbattribute.MarshalMap(record.GetData())
	if trans.err != nil {
		trans.err = fmt.Errorf(
			`failed to serialize item to put into table %q: "%w", data: %s`,
			record.GetTable(), trans.err, ss.Dump(record))
	}
	return result
}

////////////////////////////////////////////////////////////////////////////////
