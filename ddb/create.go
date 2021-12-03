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

type Create interface {
	ss.NoCopy

	Condition(string) Create
	Values(Values) Create

	Request()
	RequestConditioned() bool
}

type CreateIfNotExists interface {
	ss.NoCopy

	Request() bool
}

////////////////////////////////////////////////////////////////////////////////

func (client *client) Create(record DataRecord) Create {
	result := newCreate(record, client.db)
	result.Condition(
		fmt.Sprintf("attribute_not_exists(%s)", record.GetKeyPartitionField()))
	return result
}

func (client *client) CreateIfNotExists(record DataRecord) CreateIfNotExists {
	result := newCreateIfNotExists(record, client.db)
	result.Condition(
		fmt.Sprintf("attribute_not_exists(%s)", record.GetKeyPartitionField()))
	return result
}

func (client *client) CreateOrReplace(record DataRecord) Create {
	return newCreate(record, client.db)
}

////////////////////////////////////////////////////////////////////////////////

type create struct {
	ss.NoCopyImpl

	db    *dynamodb.DynamoDB
	input dynamodb.PutItemInput
}

func newCreate(record DataRecord, db *dynamodb.DynamoDB) *create {
	result := create{
		db: db,
		input: dynamodb.PutItemInput{
			TableName: aws.String(ss.S.NewBuildEntityName(record.GetTable())),
		},
	}
	var err error
	result.input.Item, err = dynamodbattribute.MarshalMap(record.GetData())
	if err != nil {
		ss.S.Log().Panic(
			ss.NewLogMsg(
				`failed to serialize item to put into table %q`,
				record.GetTable()).
				AddDump(result.input).
				AddErr(err))
	}
	return &result
}

func (trans *create) condition(condition string) {
	condition = "(" +
		*aliasReservedInString(condition, &trans.input.ExpressionAttributeNames) +
		")"
	if trans.input.ConditionExpression == nil {
		trans.input.ConditionExpression = &condition
	} else {
		*trans.input.ConditionExpression += " and " + condition
	}
}

func (trans *create) Condition(condition string) Create {
	trans.condition(condition)
	return trans
}

func (trans *create) values(values Values) {
	values.Marshal(&trans.input.ExpressionAttributeValues)
}

func (trans *create) Values(values Values) Create {
	trans.values(values)
	return trans
}

func (trans *create) request() error {
	request, _ := trans.db.PutItemRequest(&trans.input)
	return request.Send()
}

func (trans *create) Request() {
	if err := trans.request(); err != nil {
		ss.S.Log().Panic(
			ss.
				NewLogMsg(`failed to put item into table %q`, *trans.input.TableName).
				AddDump(trans.input).
				AddErr(err))
	}
}

func (trans *create) requestConditioned() bool {
	if err := trans.request(); err != nil {
		if IsConditionalCheckError(err) {
			return false
		}
		ss.S.Log().Panic(
			ss.
				NewLogMsg(`failed to put item into table %q`, *trans.input.TableName).
				AddDump(trans.input).
				AddErr(err))
	}
	return true
}

func (trans *create) RequestConditioned() bool {
	return trans.requestConditioned()
}

////////////////////////////////////////////////////////////////////////////////

type createIfNotExists struct{ create }

func newCreateIfNotExists(
	record DataRecord,
	db *dynamodb.DynamoDB,
) *createIfNotExists {
	return &createIfNotExists{create: *newCreate(record, db)}
}

func (trans *createIfNotExists) Request() bool {
	return trans.requestConditioned()
}

////////////////////////////////////////////////////////////////////////////////
