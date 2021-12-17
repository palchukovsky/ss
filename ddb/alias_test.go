// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ddb_test

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/golang/mock/gomock"
	"github.com/palchukovsky/ss"
	"github.com/palchukovsky/ss/ddb"
	mock_ss "github.com/palchukovsky/ss/mock"
	"github.com/stretchr/testify/assert"
)

type testAliasRecord struct{}

func newTestAliasRecord() testAliasRecord { return testAliasRecord{} }

func (testAliasRecord) GetTable() string             { return "" }
func (testAliasRecord) GetKey() interface{}          { return struct{}{} }
func (testAliasRecord) GetKeyPartitionField() string { return "" }
func (testAliasRecord) GetKeySortField() string      { return "" }

func Test_DDB_Alias_AliasReservedInString(test *testing.T) {
	mock := gomock.NewController(test)
	defer mock.Finish()
	assert := assert.New(test)

	service := mock_ss.NewMockService(mock)
	service.EXPECT().NewBuildEntityName(gomock.Any()).Times(2).Return("test_name")
	ss.Set(service)

	update := "set filed = 1, field2= 2, field3=3 remove filed4, field5"
	trans := ddb.NewWriteTrans(false)
	trans.Update(newTestAliasRecord(), update)
	input := trans.GetResult()
	assert.Equal(1, len(input.TransactItems))
	assert.NotNil(update, input.TransactItems[0].Update)
	assert.Equal(update, *input.TransactItems[0].Update.UpdateExpression)
	assert.Nil(input.TransactItems[0].Update.ExpressionAttributeNames)

	update = "set user = 1, next 2 = 2 remove snapshot, field4, share.#l and attribute_not_exists(owner)"
	trans = ddb.NewWriteTrans(false)
	trans.Update(newTestAliasRecord(), update).Alias("#l", "l")
	input = trans.GetResult()
	assert.Equal(1, len(input.TransactItems))
	assert.NotNil(update, input.TransactItems[0].Update)
	assert.Equal(
		"set #user = 1, #next 2 = 2 remove #snapshot, field4, #share.#l and attribute_not_exists(#owner)",
		*input.TransactItems[0].Update.UpdateExpression)
	assert.NotNil(input.TransactItems[0].Update.ExpressionAttributeNames)
	assert.Equal(
		map[string]*string{
			"#l":        aws.String("l"),
			"#user":     aws.String("user"),
			"#next":     aws.String("next"),
			"#snapshot": aws.String("snapshot"),
			"#owner":    aws.String("owner"),
			"#share":    aws.String("share"),
		},
		input.TransactItems[0].Update.ExpressionAttributeNames)
}
