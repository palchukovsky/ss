// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ss_test

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/palchukovsky/ss"
	"github.com/stretchr/testify/assert"
)

func Test_SS_Time_MarshalJSON(test *testing.T) {
	assert := assert.New(test)

	source := `{"val":"20090107 081258"}`

	value := struct {
		Value ss.Time `json:"val"`
	}{}
	err := json.Unmarshal([]byte(source), &value)
	assert.NoError(err)
	assert.True(
		value.Value.Equal(
			ss.NewTime(
				time.Date(2009, 1, 7, 8, 12, 58, 0, time.UTC))))
	assert.False(
		value.Value.Equal(
			ss.NewTime(
				time.Date(2009, 1, 7, 8, 12, 59, 0, time.UTC))))

	exported, err := json.Marshal(value)
	assert.NoError(err)
	assert.Equal(source, string(exported))
}

func Test_SS_DateOrTime(test *testing.T) {
	assert := assert.New(test)

	time := ss.NewTime(time.Unix(1231317945, 0))

	dateOrTime := ss.NewDateOrTime(time, false)
	assert.Equal(time.Get().Unix(), dateOrTime.Value.Get().Unix())
	assert.False(dateOrTime.IsDateOnly)
	dbVal, err := dynamodbattribute.Marshal(dateOrTime)
	assert.NoError(err)
	assert.Equal(*dbVal.N, "12313179451")

	dateOrTime = ss.DateOrTime{}
	assert.NoError(dynamodbattribute.Unmarshal(dbVal, &dateOrTime))
	assert.Equal(time.Get().Unix(), dateOrTime.Value.Get().Unix())
	assert.False(dateOrTime.IsDateOnly)

	dateOrTime = ss.NewDateOrTime(time, true)
	assert.Equal(time.Get().Unix(), dateOrTime.Value.Get().Unix())
	assert.True(dateOrTime.IsDateOnly)
	dbVal, err = dynamodbattribute.Marshal(dateOrTime)
	assert.NoError(err)
	assert.Equal(*dbVal.N, "12313179450")

	dateOrTime = ss.DateOrTime{}
	assert.NoError(dynamodbattribute.Unmarshal(dbVal, &dateOrTime))
	assert.Equal(time.Get().Unix(), dateOrTime.Value.Get().Unix())
	assert.True(dateOrTime.IsDateOnly)
}

func Test_SS_Date(test *testing.T) {
	assert := assert.New(test)
	{
		date := ss.NewDate(2009, 1, 7)
		assert.Equal(`"20090107"`, fmt.Sprintf("%q", date))
		dbVal, err := dynamodbattribute.Marshal(date)
		assert.NoError(err)
		assert.Equal(*dbVal.N, "20090107")
	}
	{
		var date ss.Date
		dbVal := dynamodb.AttributeValue{N: aws.String("19950101")}
		assert.NoError(dynamodbattribute.Unmarshal(&dbVal, &date))
		assert.Equal(`"19950101"`, fmt.Sprintf("%q", date))
	}
}
