// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ddb

import (
	"fmt"

	"github.com/aws/aws-sdk-go/service/dynamodb"
)

////////////////////////////////////////////////////////////////////////////////

// WriteTrans helps to build write-transaction.
type WriteTrans interface {
	Result() (dynamodb.TransactWriteItemsInput, error)
	IsEmpty() bool
	GetSize() int

	Create(DataRecord) CreateTrans
	CreateOrUpdate(DataRecord) CreateTrans
	Check(KeyRecord) CheckTrans
	Update(key KeyRecord, update string) UpdateTrans
	Delete(KeyRecord) DeleteTrans
}

// NewWriteTrans creates new write transaction builder.
func NewWriteTrans() WriteTrans {
	return &writeTrans{result: []*dynamodb.TransactWriteItem{}}
}

// WriteTransExpression describes the part of WriteTrans
// witch builds typed expression.
type WriteTransExpression interface{}

////////////////////////////////////////////////////////////////////////////////

type writeTrans struct {
	err    error
	result []*dynamodb.TransactWriteItem
}

func (trans writeTrans) Result() (dynamodb.TransactWriteItemsInput, error) {
	return dynamodb.TransactWriteItemsInput{TransactItems: trans.result},
		trans.err
}

func (trans writeTrans) IsEmpty() bool { return len(trans.result) == 0 }
func (trans writeTrans) GetSize() int  { return len(trans.result) }

////////////////////////////////////////////////////////////////////////////////

type writeTransExpression struct{ trans *writeTrans }

func newWriteTransExpression(
	trans *writeTrans,
	result dynamodb.TransactWriteItem,
) writeTransExpression {
	trans.result = append(trans.result, &result)
	return writeTransExpression{trans: trans}
}

func (trans writeTransExpression) marshalValues(values Values,
) map[string]*dynamodb.AttributeValue {
	if trans.trans.err != nil {
		return nil
	}
	var result map[string]*dynamodb.AttributeValue
	result, trans.trans.err = values.Marshal()
	if trans.trans.err != nil {
		trans.trans.err = fmt.Errorf(
			`failed to serialize values: "%w"`,
			trans.trans.err)
	}
	return result
}

////////////////////////////////////////////////////////////////////////////////
