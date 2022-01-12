// Copyright 2021-2022, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ss_test

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/google/uuid"
	"github.com/palchukovsky/ss"
	"github.com/stretchr/testify/assert"
)

const (
	entityIDSource = "dl8e8DgqQfu5fspTVMp3Gg"
	entityUUID     = "765f1ef0-382a-41fb-b97e-ca5354ca771a"
)

func Test_SS_EntityID_Export(test *testing.T) {
	assert := assert.New(test)

	id, err := ss.ParseEntityID(entityIDSource)
	assert.NoError(err)
	assert.NotNil(id)

	uuid, err := uuid.Parse(entityUUID)
	assert.NoError(err)
	assert.Equal(entityIDSource, base64.RawStdEncoding.EncodeToString(uuid[:]))

	assert.Equal(entityIDSource, id.String())
	assert.Equal(uuid[:], id[:])
	assert.Equal(fmt.Sprintf("|%s|", entityIDSource), fmt.Sprintf("|%s|", id))
}

func Test_SS_EntityID_ParseError(test *testing.T) {
	assert := assert.New(test)
	{
		id, err := ss.ParseEntityID("")
		assert.EqualError(err,
			`failed to unmarshal entity ID: "invalid UUID (got 0 bytes)"`)
		assert.Equal(make([]byte, 16), []byte(id[:]))
	}
	{
		/* spell-checker: disable */
		id, err := ss.ParseEntityID("AQIDBAUGBwgJCgsMDQ4PEBE")
		/* spell-checker: enable */
		assert.EqualError(err,
			`failed to unmarshal entity ID: "invalid UUID (got 17 bytes)"`)
		assert.Equal(make([]byte, 16), []byte(id[:]))
	}
	{
		id, err := ss.ParseEntityID("dl8e8Dgq5fspTVMp3Gg==")
		assert.EqualError(err,
			`failed to decode entity ID: "illegal base64 data at input byte 19"`)
		assert.Equal(make([]byte, 16), []byte(id[:]))
	}
}

func Test_SS_EntityID_MarshalJSON(test *testing.T) {
	assert := assert.New(test)

	id, err := ss.ParseEntityID(entityIDSource)
	assert.NoError(err)

	value, err := json.Marshal(id)
	assert.NoError(err)
	assert.Equal(fmt.Sprintf("%q", entityIDSource), string(value))

	restoredValue := &ss.EntityID{}
	assert.NoError(json.Unmarshal(value, restoredValue))
	assert.Equal(entityIDSource, restoredValue.String())

	record := struct {
		Value ss.EntityID `json:"value"`
	}{Value: id}
	recordValue, err := json.Marshal(record)
	assert.NoError(err)
	assert.Equal(fmt.Sprintf(`{"value":%q}`, entityIDSource),
		string(recordValue))

	restoredRecord := &struct {
		Value ss.EntityID `json:"value"`
	}{}
	assert.NoError(json.Unmarshal(recordValue, restoredRecord))
	assert.Equal(entityIDSource, restoredRecord.Value.String())
}

func Test_SS_EntityID_MarshalDynamoDB(test *testing.T) {
	assert := assert.New(test)

	id, err := ss.ParseEntityID(entityIDSource)
	assert.NoError(err)
	assert.NotNil(id)
	control, err := uuid.Parse(entityUUID)
	assert.NoError(err)

	ssValue, err := dynamodbattribute.Marshal(id)
	assert.NoError(err)
	controlValue, err := dynamodbattribute.Marshal(control)
	assert.NoError(err)
	assert.Equal(*controlValue, *ssValue)

	restoredSSValue := &ss.EntityID{}
	assert.NoError(dynamodbattribute.Unmarshal(controlValue, restoredSSValue))
	assert.Equal(entityIDSource, restoredSSValue.String())

	ssRecord := struct {
		Value ss.EntityID `json:"value"`
	}{Value: id}
	controlRecord := struct {
		Value uuid.UUID `json:"value"`
	}{Value: control}
	ssRecordValue, err := dynamodbattribute.MarshalMap(ssRecord)
	assert.NoError(err)
	controlRecordValue, err := dynamodbattribute.MarshalMap(controlRecord)
	assert.NoError(err)
	assert.Equal(controlRecordValue, ssRecordValue)

	restoredSSRecord := &struct {
		Value ss.EntityID `json:"value"`
	}{}
	assert.NoError(
		dynamodbattribute.UnmarshalMap(controlRecordValue, restoredSSRecord))
	assert.Equal(entityIDSource, restoredSSRecord.Value.String())
}
