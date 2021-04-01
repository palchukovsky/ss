// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ddb

import (
	"fmt"

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
	Err() error
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
	GetAt(index int) (RecordBuffer, error)
}

////////////////////////////////////////////////////////////////////////////////

func newCacheIterator(
	data []map[string]*dynamodb.AttributeValue,
	record RecordBuffer,
) CacheIterator {
	return &cacheIterator{data: data, record: record}
}

type cacheIterator struct {
	data   []map[string]*dynamodb.AttributeValue
	record RecordBuffer
	pos    int
	err    error
}

func (it cacheIterator) GetSize() int      { return len(it.data) }
func (it cacheIterator) Err() error        { return it.err }
func (it cacheIterator) Get() RecordBuffer { return it.record }

func (it *cacheIterator) Next() bool {
	if it.err != nil {
		return false
	}
	if it.pos >= len(it.data) {
		return false
	}
	it.pos++
	it.readAt(it.pos - 1)
	return it.err == nil
}

func (it cacheIterator) GetAt(index int) (RecordBuffer, error) {
	if it.err != nil {
		return nil, it.err
	}
	it.readAt(index)
	return it.record, it.err
}

func (it *cacheIterator) readAt(index int) {
	it.record.Clear()
	it.err = dynamodbattribute.UnmarshalMap(it.data[index], it.record)
	if it.err != nil {
		it.err = fmt.Errorf(
			`failed to unmarshal cached row %d of %d`+
				` from with type %q from row set: "%w"`,
			index,
			len(it.data),
			ss.GetTypeName(it.record),
			it.err)
	}
}

////////////////////////////////////////////////////////////////////////////////

func newPagedIterator(request *awsrequest.Request, record RecordBuffer,
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
	err       error
}

func (it pagedIterator) Err() error        { return it.err }
func (it pagedIterator) Get() RecordBuffer { return it.cache.Get() }

func (it *pagedIterator) Next() bool {
	it.err = nil
	if it.cache.Next() {
		return true
	}
	if it.err = it.cache.Err(); it.err != nil {
		return false
	}
	// example has been taken from func (DynamoDB) QueryPagesWithContext:
	if !it.paginator.Next() {
		if err := it.paginator.Err(); err != nil {
			it.err = fmt.Errorf(`failed to request next page "%w"`, err)
		}
		return false
	}
	it.cache = newCacheIterator(
		it.paginator.Page().(*dynamodb.QueryOutput).Items,
		it.cache.Get())
	return it.cache.Next()
}

////////////////////////////////////////////////////////////////////////////////
