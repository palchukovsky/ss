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

// Client is a database client interface.
type Client interface {
	Index(resultRecord IndexRecord) Index

	Get(KeyRecordBuffer) Get
	Find(KeyRecordBuffer) Find
	Query(record RecordBuffer, keyCondition string, values Values) Query

	Create(data DataRecord) error
	CreateIfNotExists(data DataRecord) (bool, error)
	CreateOrUpdate(data DataRecord) error
	Update(key KeyRecord, update string) Update
	Delete(key KeyRecord) Delete

	Write(WriteTrans) error
}

// GetClientInstance returns reference to client singleton.
func GetClientInstance() Client {
	if clientInstance == nil {
		clientInstance = client{db: dynamodb.New(ss.S.NewAWSSessionV1())}
	}
	return clientInstance
}

// Index describes db-command interface for the table index.
type Index interface {
	Query(keyCondition string, values Values) Query
}

////////////////////////////////////////////////////////////////////////////////

var clientInstance Client

type client struct{ db *dynamodb.DynamoDB }

func (client client) DynamoDB() *dynamodb.DynamoDB { return client.db }

func (client client) Index(record IndexRecord) Index {
	return index{client: client, record: record}
}

func (client client) Find(key KeyRecordBuffer) Find {
	return newFind(client.db, key)
}
func (client client) Get(key KeyRecordBuffer) Get {
	return newGet(client.db, key)
}

func (client client) Create(record DataRecord) error {
	input, err := client.newPut(record)
	if err != nil {
		return err
	}
	input.ConditionExpression = aws.String(fmt.Sprintf(
		"attribute_not_exists(%s)", aliasReservedWord(
			record.GetKeyPartitionField(), &input.ExpressionAttributeNames)))
	return client.put(input)
}

func (client client) CreateIfNotExists(record DataRecord) (bool, error) {
	err := client.Create(record)
	if err == nil {
		return true, nil
	}
	if IsConditionalCheckError(err) {
		return false, nil
	}
	return false, err
}

func (client client) CreateOrUpdate(record DataRecord) error {
	input, err := client.newPut(record)
	if err != nil {
		return err
	}
	return client.put(input)
}

func (client client) newPut(record DataRecord) (dynamodb.PutItemInput, error) {
	result := dynamodb.PutItemInput{
		TableName: aws.String(ss.S.NewBuildEntityName(record.GetTable())),
	}
	var err error
	result.Item, err = dynamodbattribute.MarshalMap(record.GetData())
	if err != nil {
		return result, fmt.Errorf(
			`failed to serialize item to put into table %q: "%w", data: %s`,
			record.GetTable(), err, ss.Dump(record))
	}
	return result, nil
}

func (client client) put(input dynamodb.PutItemInput) error {
	request, _ := client.db.PutItemRequest(&input)
	if err := request.Send(); err != nil {
		return fmt.Errorf(`failed to put item into table %q: "%w", input: %s`,
			*input.TableName, err, ss.Dump(input))
	}
	return nil
}

func (client client) Write(trans WriteTrans) error {
	input, err := trans.Result()
	if err != nil {
		return err
	}
	request, _ := client.db.TransactWriteItemsRequest(&input)
	if err := request.Send(); err != nil {
		return fmt.Errorf(`failed to write trans: "%w", input dump: %s`,
			err, ss.Dump(input))
	}
	return nil
}

////////////////////////////////////////////////////////////////////////////////

type index struct {
	client client
	record IndexRecord
}

func (index index) Query(keyCondition string, values Values) Query {
	result := newQuery(index.client, index.record, keyCondition, values)
	result.input.IndexName = aws.String(index.record.GetIndex())
	return result
}

////////////////////////////////////////////////////////////////////////////////
