// Copyright 2021-2022, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ddb

import (
	"github.com/aws/aws-sdk-go/aws"
	awsrequest "github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/palchukovsky/ss"
)

////////////////////////////////////////////////////////////////////////////////

// IteratorMover describes intreface to move iterator.
type IteratorMover interface {
	Next() bool
}

// Iterator describes intreface to read paged data from the database.
type Iterator interface {
	IteratorMover
	Get() RecordBuffer
}

// CacheIterator describes intreface to read query or scan cached result.
type CacheIterator interface {
	Iterator
	GetSize() int
	GetAt(index int) RecordBuffer
}

////////////////////////////////////////////////////////////////////////////////

func newCacheIterator(
	data []map[string]*dynamodb.AttributeValue,
	record RecordBuffer,
) *cacheIterator {
	return &cacheIterator{data: data, record: record}
}

type cacheIterator struct {
	data   []map[string]*dynamodb.AttributeValue
	record RecordBuffer
	pos    int
}

func (it *cacheIterator) Set(data []map[string]*dynamodb.AttributeValue) {
	it.data = data
}

func (it cacheIterator) GetSize() int      { return len(it.data) }
func (it cacheIterator) Get() RecordBuffer { return it.record }

func (it *cacheIterator) Next() bool {
	if it.data == nil {
		ss.S.Log().Panic(ss.NewLogMsg("iterator is not initialized"))
	}
	if it.pos >= len(it.data) {
		return false
	}
	it.pos++
	it.readAt(it.pos - 1)
	return true
}

func (it cacheIterator) GetAt(index int) RecordBuffer {
	it.readAt(index)
	return it.record
}

func (it *cacheIterator) readAt(index int) {
	it.record.Clear()
	err := dynamodbattribute.UnmarshalMap(it.data[index], it.record)
	if err != nil {
		ss.S.Log().Panic(
			ss.
				NewLogMsg(
					`failed to unmarshal cached row %d of %d from row set`,
					index,
					len(it.data)).
				AddErr(err).
				AddDump(it.data).
				AddDump(it.record))
	}
}

////////////////////////////////////////////////////////////////////////////////

func newPagedIterator(
	request *awsrequest.Request,
	record RecordBuffer,
) Iterator {
	return &pagedIterator{
		// example has been taken from func (DynamoDB) QueryPagesWithContext:
		paginator: awsrequest.Pagination{
			NewRequest: func() (*awsrequest.Request, error) {
				request.SetContext(aws.BackgroundContext())
				return request, nil
			},
		},
		cache: newCacheIterator([]map[string]*dynamodb.AttributeValue{}, record),
	}
}

type pagedIterator struct {
	paginator awsrequest.Pagination
	cache     CacheIterator
}

func (it pagedIterator) Get() RecordBuffer { return it.cache.Get() }

func (it *pagedIterator) Next() bool {
	if it.cache.Next() {
		return true
	}
	// example has been taken from func (DynamoDB) QueryPagesWithContext:
	if !it.paginator.Next() {
		if err := it.paginator.Err(); err != nil {
			ss.S.Log().Panic(
				ss.NewLogMsg(`failed to request next page`).AddErr(err))
		}
		return false
	}
	it.cache = newCacheIterator(
		it.paginator.Page().(*dynamodb.QueryOutput).Items,
		it.cache.Get())
	return it.cache.Next()
}

////////////////////////////////////////////////////////////////////////////////
