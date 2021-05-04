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

// Update describes the interface to update one record by key.
type Update interface {
	Values(Values) Update
	Condition(string) Update
	Request() (bool, error)
	RequestAndReturn(RecordBuffer) (bool, error)
}

////////////////////////////////////////////////////////////////////////////////

func (client client) Update(key KeyRecord, updateExp string) Update {
	result := newUpdateTemplate(client.db, key, updateExp)
	if result.err != nil {
		return result
	}
	if result.err = result.SetKey(key.GetKey()); result.err != nil {
		return result
	}
	return result
}

func newUpdateTemplate(db *dynamodb.DynamoDB, record Record, updateExp string,
) update {
	result := update{
		db: db,
		input: dynamodb.UpdateItemInput{
			TableName:        aws.String(ss.S.NewBuildEntityName(record.GetTable())),
			UpdateExpression: aws.String(updateExp),
		},
	}
	result.input.UpdateExpression = aliasReservedInString(updateExp,
		&result.input.ExpressionAttributeNames)
	result.input.ConditionExpression = aws.String(fmt.Sprintf(
		"attribute_exists(%s)", aliasReservedWord(
			record.GetKeyPartitionField(), &result.input.ExpressionAttributeNames)))
	return result
}

type update struct {
	db    *dynamodb.DynamoDB
	input dynamodb.UpdateItemInput
	err   error
}

func (update update) Values(values Values) Update {
	if update.err != nil {
		return update
	}
	update.input.ExpressionAttributeValues, update.err = values.Marshal()
	if update.err != nil {
		update.err = fmt.Errorf(
			`failed to serialize values to update table %q: "%w", values: "%v"`,
			update.getTable(), update.err, values)
	}
	return update
}

func (update update) Condition(condition string) Update {
	*update.input.ConditionExpression += " and " +
		*aliasReservedInString(condition, &update.input.ExpressionAttributeNames)
	return update
}

func (update update) Request() (bool, error) {
	output, err := update.request()
	return output != nil, err
}

func (update update) RequestAndReturn(result RecordBuffer) (bool, error) {
	update.input.ReturnValues = aws.String(dynamodb.ReturnValueAllNew)
	output, err := update.request()
	if err != nil || output == nil {
		return false, err
	}
	err = dynamodbattribute.UnmarshalMap(output.Attributes, result)
	if err != nil {
		return false,
			fmt.Errorf(
				`failed to read update response from table %q: "%w"`,
				update.getTable(),
				err)
	}
	return true, nil
}

func (update update) request() (*dynamodb.UpdateItemOutput, error) {
	if update.err != nil {
		return nil, update.err
	}
	request, result := update.db.UpdateItemRequest(&update.input)
	if err := request.Send(); err != nil {
		if IsConditionalCheckError(err) {
			// no error, but not found
			return nil, nil
		}
		return nil,
			fmt.Errorf(
				`failed to update item in table %q: "%w"`,
				update.getTable(),
				err)
	}
	return result, nil
}

func (update *update) SetKey(source interface{}) error {
	key, err := dynamodbattribute.MarshalMap(source)
	if err != nil {
		return fmt.Errorf(
			`failed to serialize key to update table %q: "%w"`,
			update.getTable(),
			err)
	}
	update.input.Key = key
	return nil
}

func (update update) getTable() string { return *update.input.TableName }
