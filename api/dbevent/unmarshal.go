// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package apidbevent

import (
	"encoding/json"
	"fmt"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/dynamodb/dynamodbattribute"
	"github.com/palchukovsky/ss"
)

// UnmarshalEventsDynamoDBAttributeValues unmarshals db-event.
func UnmarshalEventsDynamoDBAttributeValues(
	source map[string]events.DynamoDBAttributeValue,
	result interface{},
) error {

	attrs := make(map[string]*dynamodb.AttributeValue)
	for k, v := range source {
		var attr dynamodb.AttributeValue
		bytes, err := v.MarshalJSON()
		if err != nil {
			return fmt.Errorf(
				`failed to convert from events-value to attribute: "%w"`, err)
		}
		if err := json.Unmarshal(bytes, &attr); err != nil {
			return fmt.Errorf(
				`failed to unmarshal events-value JSON into attribute: "%w"`, err)
		}
		attrs[k] = &attr
	}

	if err := dynamodbattribute.UnmarshalMap(attrs, result); err != nil {
		return fmt.Errorf(
			`failed to unmarshal events DynamoDB attribute values: "%w", dump: %v`,
			err, ss.Dump(source))
	}

	return nil
}
