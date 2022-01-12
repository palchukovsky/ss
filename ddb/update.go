// Copyright 2021-2022, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ddb

import (
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/palchukovsky/ss"
)

// Update describes the interface to update one record by key.
type Update interface {
	CheckedExpression

	Set(expression string) Update
	Remove(fieldName string) Update
	Expression(expression string) Update

	Values(Values) Update
	Value(name string, value interface{}) Update

	Alias(name, value string) Update

	Condition(string) Update

	RequestWithResult() Result
	RequestAndReturn(RecordBuffer) Result
}

////////////////////////////////////////////////////////////////////////////////

func (client *client) Update(key KeyRecord) Update {
	result := newUpdateTemplate(client.db, key)
	result.SetKey(key.GetKey())
	return result
}

func newUpdateTemplate(
	db *dynamodb.DynamoDB,
	record Record,
) *update {
	result := update{
		checkedExpression: newCheckedExpression(),
		db:                db,
		Input: dynamodb.UpdateItemInput{
			TableName: aws.String(ss.S.NewBuildEntityName(record.GetTable())),
		},
	}
	result.Input.ConditionExpression = aws.String(
		fmt.Sprintf(
			"attribute_exists(%s)",
			aliasReservedWord(
				record.GetKeyPartitionField(),
				&result.Input.ExpressionAttributeNames)))
	return &result
}

type update struct {
	checkedExpression

	db      *dynamodb.DynamoDB       `json:"-"`
	Input   dynamodb.UpdateItemInput `json:"input"`
	Expr    string                   `json:"expression"`
	Sets    []string                 `json:"sets"`
	Removes []string                 `json:"removes"`
}

func (update *update) Set(expression string) Update {
	update.Sets = append(update.Sets, expression)
	return update
}

func (update *update) Remove(fieldName string) Update {
	update.Removes = append(update.Removes, fieldName)
	return update
}

func (update *update) Expression(expression string) Update {
	if update.Expr != "" {
		update.Expr += " "
	}
	update.Expr += expression
	return update
}

func (update *update) Values(values Values) Update {
	values.Marshal(&update.Input.ExpressionAttributeValues)
	return update
}

func (update *update) Value(name string, value interface{}) Update {
	return update.Values(Values{name: value})
}

func (update *update) Alias(name, value string) Update {
	if update.Input.ExpressionAttributeNames == nil {
		update.Input.ExpressionAttributeNames = map[string]*string{
			name: aws.String(value),
		}
	} else {
		update.Input.ExpressionAttributeNames[name] = aws.String(value)
	}
	return update
}

func (update *update) Condition(condition string) Update {
	*update.Input.ConditionExpression += " and (" +
		*aliasReservedInString(condition, &update.Input.ExpressionAttributeNames) +
		")"
	return update
}

func (update *update) RequestWithResult() Result {
	result, _ := update.request()
	return result
}

func (update *update) RequestAndReturn(resultRecord RecordBuffer) Result {
	update.Input.ReturnValues = aws.String(dynamodb.ReturnValueAllNew)
	result, output := update.request()
	if !result.IsSuccess() {
		return result
	}
	err := dynamodbattribute.UnmarshalMap(output.Attributes, resultRecord)
	if err != nil {
		ss.S.Log().Panic(
			ss.
				NewLogMsg(
					"failed to read update response from table %q",
					update.getTable()).
				AddErr(err))
	}
	return result
}

func (update *update) request() (Result, *dynamodb.UpdateItemOutput) {
	{
		expression := make([]string, 0, 3)
		if update.Expr != "" {
			expression = append(expression, update.Expr)
		}
		if sets := strings.Join(update.Sets, ","); sets != "" {
			expression = append(expression, "set "+sets)
		}
		if removes := strings.Join(update.Removes, ","); removes != "" {
			expression = append(expression, "remove "+removes)
		}
		update.Input.UpdateExpression = aliasReservedInString(
			strings.Join(expression, " "),
			&update.Input.ExpressionAttributeNames)
	}
	request, output := update.db.UpdateItemRequest(&update.Input)
	result, err := newResult(request.Send(), update.isConditionalCheckFailAllowed)
	if err != nil {
		ss.S.Log().Panic(
			ss.
				NewLogMsg("failed to update item in table %q", update.getTable()).
				AddDump(update).
				AddErr(err))
	}
	return result, output
}

func (update *update) SetKey(source interface{}) {
	key, err := dynamodbattribute.MarshalMap(source)
	if err != nil {
		ss.S.Log().Panic(
			ss.
				NewLogMsg(
					`failed to serialize key to update table %q`,
					update.getTable()).
				AddErr(err))
	}
	update.Input.Key = key
}

func (update *update) getTable() string { return *update.Input.TableName }
