// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ddb

import (
	"fmt"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/palchukovsky/ss"
)

// Query describes the interface to query records from a database.
type Query interface {
	Filter(string) Query
	Limit(int64) Query
	Descending() Query
	RequestOne() (bool, error)
	RequestPaged() (Iterator, error)
	RequestAll() (CacheIterator, error)
	Update(update string, values Values) (uint, uint, error)
	UpdateWithCondition(
		update string, values Values, prepareUpdate func(Update) Update,
	) (uint, uint, error)
	Trans(newTrans func() WriteTrans) (uint, uint, error)
	TransAndCheckError(
		newTrans func() WriteTrans,
		checkErrOrRepeat func(error) (bool, error),
	) (uint, uint, error)
}

////////////////////////////////////////////////////////////////////////////////

func (client client) Query(
	record RecordBuffer,
	keyCondition string,
	values Values,
) Query {
	return newQuery(client, record, keyCondition, values)
}

func newQuery(
	client client,
	record RecordBuffer,
	keyCondition string,
	values Values,
) query {
	result := query{
		client: client,
		record: record,
		input: dynamodb.QueryInput{
			TableName: aws.String(ss.S.NewBuildEntityName(record.GetTable())),
		},
	}
	result.input.ExpressionAttributeValues, result.err = values.Marshal()
	if result.err != nil {
		result.err = fmt.Errorf(
			`failed to serialize values to query from table %q: "%w", values: "%v"`,
			result.record.GetTable(), result.err, values)
		return result
	}

	result.input.KeyConditionExpression = aliasReservedInString(keyCondition,
		&result.input.ExpressionAttributeNames)
	result.input.ProjectionExpression = getRecordProjection(record,
		&result.input.ExpressionAttributeNames)

	return result
}

type query struct {
	client client
	record RecordBuffer
	input  dynamodb.QueryInput
	err    error
}

func (query query) Filter(filter string) Query {
	query.input.FilterExpression = aliasReservedInString(filter,
		&query.input.ExpressionAttributeNames)
	return query
}

func (query query) Limit(limit int64) Query {
	query.input.Limit = &limit
	return query
}

func (query query) Descending() Query {
	query.input.ScanIndexForward = ss.BoolPtr(false)
	return query
}

func (query query) RequestPaged() (Iterator, error) {
	if query.err != nil {
		return nil, query.err
	}
	request, _ := query.client.db.QueryRequest(&query.input)
	return newPagedIterator(request, query.record), nil
}

func (query query) RequestOne() (bool, error) {
	it, err := query.RequestAll()
	if err != nil {
		return false, err
	}
	if it.GetSize() > 1 {
		return false, fmt.Errorf(`expected one record, but returned %d`,
			it.GetSize())
	}
	isFound := it.Next()
	return isFound, it.Err()
}

func (query query) RequestAll() (CacheIterator, error) {
	if query.err != nil {
		return nil, query.err
	}
	request, output := query.client.db.QueryRequest(&query.input)
	if err := request.Send(); err != nil {
		return nil, fmt.Errorf(`failed to query from table %q: "%w", input: %v`,
			query.record.GetTable(), err, query.input)
	}
	return newCacheIterator(output.Items, query.record), nil
}

func (query query) Update(updateExp string, values Values) (uint, uint, error) {
	update := newUpdateTemplate(query.client.db, query.record, updateExp).
		Values(values).(update)
	return query.update(update)
}
func (query query) UpdateWithCondition(
	updateExp string,
	values Values,
	prepareUpdate func(Update) Update,
) (uint,
	uint,
	error,
) {
	update := prepareUpdate(
		newUpdateTemplate(query.client.db, query.record, updateExp),
	).Values(values).(update)
	return query.update(update)
}

func (query query) update(update update) (uint, uint, error) {
	it, err := query.RequestPaged()
	if err != nil {
		return 0, 0, err
	}

	var updatedCount uint
	var querySize uint
	for it.Next() {
		querySize++
		// Clone object here if it became parallel, or remove this event and user
		// only Trans.
		if err := update.SetKey(it.Get()); err != nil {
			return querySize, updatedCount, err
		}
		isFound, err := update.Request()
		if err != nil {
			return querySize, updatedCount, err
		}
		if isFound {
			updatedCount++
		}
	}
	return querySize, updatedCount, it.Err()
}

func (query query) Trans(newTrans func() WriteTrans) (uint, uint, error) {
	return query.TransAndCheckError(
		newTrans, func(err error) (bool, error) { return false, err })
}

func (query query) TransAndCheckError(
	newTrans func() WriteTrans,
	checkErrOrRepeat func(error) (bool, error),
) (
	querySize uint,
	updatedCount uint,
	err error,
) {

	it, err := query.RequestPaged()
	if err != nil {
		return
	}

	for it.Next() {
		querySize++
		attempt := 1
		for {
			trans := newTrans()
			err := query.client.Write(trans)
			if err == nil {
				updatedCount++
				break
			}
			repeat, err := checkErrOrRepeat(err)
			if err != nil {
				return querySize,
					updatedCount,
					fmt.Errorf(
						`failed to execute trans for queried record %d (%d) on attempt %d: "%w"`,
						updatedCount,
						querySize,
						attempt,
						err)
			}
			if !repeat {
				break
			}
			if attempt <= 10 {
				return querySize,
					updatedCount,
					fmt.Errorf(
						`failed to execute atomic trans for queried record %d (%d) %d times "%w"`,
						updatedCount,
						querySize,
						attempt,
						err)
			}
			attempt++
		}
	}

	err = it.Err()

	return
}
