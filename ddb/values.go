// Copyright 2021-2022, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ddb

import (
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/palchukovsky/ss"
)

// Values is an abstract value set type.
type Values map[string]interface{}

// Marshal converts values into Dynamodb values format.
func (values Values) Marshal(dest *map[string]*dynamodb.AttributeValue) {
	result, err := dynamodbattribute.MarshalMap(values)
	if err != nil {
		ss.S.Log().Panic(
			ss.
				NewLogMsg(`failed to serialize DDB request values`).
				AddErr(err).
				AddDump(values))
	}
	if *dest == nil {
		*dest = result
		return
	}
	for k, v := range result {
		(*dest)[k] = v
	}
}
