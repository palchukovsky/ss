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

// Delete describes the interface to delete one record by key.
type Delete interface {
	Condition(string) Delete
	Values(Values) Delete
	Request() (bool, error)
	RequestAndReturn(RecordBuffer) (bool, error)
}

////////////////////////////////////////////////////////////////////////////////

func (client client) Delete(key KeyRecord) Delete {
	result := delete{
		db: client.db,
		input: dynamodb.DeleteItemInput{
			TableName: aws.String(ss.S.NewBuildEntityName(key.GetTable())),
		},
	}
	result.input.Key, result.err = dynamodbattribute.MarshalMap(key.GetKey())
	if result.err != nil {
		result.err = fmt.Errorf(
			`failed to serialize key to delete from table %q: "%w", key: %s`,
			result.getTable(), result.err, ss.Dump(key))
	}
	result.input.ConditionExpression = aws.String(fmt.Sprintf(
		"attribute_exists(%s)", aliasReservedWord(
			key.GetKeyPartitionField(), &result.input.ExpressionAttributeNames)))
	return result
}

type delete struct {
	db    *dynamodb.DynamoDB
	input dynamodb.DeleteItemInput
	err   error
}

func (delete delete) Values(values Values) Delete {
	if delete.err != nil {
		return delete
	}
	delete.input.ExpressionAttributeValues, delete.err = values.Marshal()
	if delete.err != nil {
		delete.err = fmt.Errorf(
			`failed to serialize values to update table %q: "%w", values: "%v"`,
			delete.getTable(), delete.err, values)
	}
	return delete
}

func (delete delete) Condition(condition string) Delete {
	*delete.input.ConditionExpression += " and " +
		*aliasReservedInString(condition, &delete.input.ExpressionAttributeNames)
	return delete
}

func (delete delete) Request() (bool, error) {
	output, err := delete.request()
	return output != nil, err
}

func (delete delete) RequestAndReturn(result RecordBuffer) (bool, error) {
	delete.input.ReturnValues = aws.String(dynamodb.ReturnValueAllOld)
	output, err := delete.request()
	if err != nil || output == nil {
		return false, err
	}
	err = dynamodbattribute.UnmarshalMap(output.Attributes, result)
	if err != nil {
		return false, fmt.Errorf(
			`failed to read delete response from table %q: "%w", dump: %s`,
			delete.getTable(), err, ss.Dump(output.Attributes))
	}
	return true, nil
}

func (delete delete) request() (*dynamodb.DeleteItemOutput, error) {
	if delete.err != nil {
		return nil, delete.err
	}
	request, result := delete.db.DeleteItemRequest(&delete.input)
	if err := request.Send(); err != nil {
		if delete.input.ConditionExpression != nil {
			if IsConditionalCheckError(err) {
				// no error, but not found
				return nil, nil
			}
		}
		return nil, fmt.Errorf(
			`failed to delete item from table %q: "%w", input: %s`,
			delete.getTable(), err, ss.Dump(delete.input))
	}
	return result, nil
}

func (delete delete) getTable() string { return *delete.input.TableName }
