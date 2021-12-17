// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ddb

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/palchukovsky/ss"
)

////////////////////////////////////////////////////////////////////////////////

// WriteTrans helps to build write-transaction.
type WriteTrans interface {
	ss.NoCopy

	IsEmpty() bool
	GetSize() int

	CreateIfNotExists(DataRecord) CreateTrans
	CreateOrReplace(DataRecord) CreateTrans
	Replace(DataRecord) CreateTrans
	Check(KeyRecord) CheckTrans
	Update(key KeyRecord, update string) UpdateTrans
	Delete(KeyRecord) DeleteTrans
	DeleteIfExisting(KeyRecord) DeleteTrans

	MarshalLogMsg(destination map[string]interface{})

	GetResult() *dynamodb.TransactWriteItemsInput
	getAllowedToFailConditionalChecks() map[int]struct{}
}

// NewWriteTrans creates new write transaction builder.
func NewWriteTrans() WriteTrans {
	return &writeTrans{result: []*dynamodb.TransactWriteItem{}}
}

// WriteTransExpression describes the part of WriteTrans
// witch builds typed expression.
type WriteTransExpression interface{ CheckedTransExpression }

////////////////////////////////////////////////////////////////////////////////

type writeTrans struct {
	ss.NoCopyImpl

	result                         []*dynamodb.TransactWriteItem
	allowedToFailConditionalChecks map[int]struct{}
}

func (trans *writeTrans) GetResult() *dynamodb.TransactWriteItemsInput {
	return &dynamodb.TransactWriteItemsInput{TransactItems: trans.result}
}

func (trans *writeTrans) getAllowedToFailConditionalChecks() map[int]struct{} {
	return trans.allowedToFailConditionalChecks
}

func (trans *writeTrans) IsEmpty() bool { return len(trans.result) == 0 }
func (trans *writeTrans) GetSize() int  { return len(trans.result) }

func (trans *writeTrans) MarshalLogMsg(destination map[string]interface{}) {
	ss.MarshalLogMsgAttrDump(trans.result, destination)
}

////////////////////////////////////////////////////////////////////////////////

type writeTransExpression struct {
	checkedTransExpression
	trans *writeTrans
}

func newWriteTransExpression(
	trans *writeTrans,
	result dynamodb.TransactWriteItem,
) writeTransExpression {
	trans.result = append(trans.result, &result)
	return writeTransExpression{
		checkedTransExpression: newCheckedTransExpression(
			len(trans.result)-1,
			trans.getAllowedToFailConditionalChecks()),
		trans: trans,
	}
}

func (trans *writeTransExpression) marshalValues(
	source Values,
	destination *map[string]*dynamodb.AttributeValue,
) {
	source.Marshal(destination)
}

func (*writeTransExpression) addAlias(
	name string,
	value string,
	dest *map[string]*string,
) {
	if *dest == nil {
		*dest = map[string]*string{name: aws.String(value)}
	} else {
		(*dest)[name] = aws.String(value)
	}
}

////////////////////////////////////////////////////////////////////////////////
