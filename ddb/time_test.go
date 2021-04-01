// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ddb_test

import (
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/palchukovsky/ss/ddb"
	"github.com/stretchr/testify/assert"
)

func Test_DDB_DateOrTime(test *testing.T) {
	assert := assert.New(test)

	time := ddb.NewTime(time.Unix(1231317945, 0))

	dateOrTime := ddb.NewDateOrTime(time, false)
	assert.Equal(time.Get().Unix(), dateOrTime.Value.Get().Unix())
	assert.False(dateOrTime.IsDateOnly)
	dbVal, err := dynamodbattribute.Marshal(dateOrTime)
	assert.NoError(err)
	assert.Equal(*dbVal.N, "12313179450")

	dateOrTime = ddb.DateOrTime{}
	assert.NoError(dynamodbattribute.Unmarshal(dbVal, &dateOrTime))
	assert.Equal(time.Get().Unix(), dateOrTime.Value.Get().Unix())
	assert.False(dateOrTime.IsDateOnly)

	dateOrTime = ddb.NewDateOrTime(time, true)
	assert.Equal(time.Get().Unix(), dateOrTime.Value.Get().Unix())
	assert.True(dateOrTime.IsDateOnly)
	dbVal, err = dynamodbattribute.Marshal(dateOrTime)
	assert.NoError(err)
	assert.Equal(*dbVal.N, "12313179451")

	dateOrTime = ddb.DateOrTime{}
	assert.NoError(dynamodbattribute.Unmarshal(dbVal, &dateOrTime))
	assert.Equal(time.Get().Unix(), dateOrTime.Value.Get().Unix())
	assert.True(dateOrTime.IsDateOnly)
}
