// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ddb

import (
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
)

// Values is an abstract value set type.
type Values map[string]interface{}

// Marshal converts values into Dynamodb values format.
func (values Values) Marshal() (map[string]*dynamodb.AttributeValue, error) {
	result, err := dynamodbattribute.MarshalMap(values)
	if err != nil {
		return nil, err
	}
	return result, nil
}
