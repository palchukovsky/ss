// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ddb

import (
	"strconv"
	stdtime "time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/palchukovsky/ss"
)

////////////////////////////////////////////////////////////////////////////////

// Time is time type for the project database.
type Time stdtime.Time

// Now returns current time in project database format.
func Now() Time { return Time(ss.Now()) }

// NewTime creates time object.
func NewTime(source stdtime.Time) Time { return Time(source) }

// Add return time with added duration. Use -duration to remove duration.
func (time Time) Add(duration stdtime.Duration) Time {
	return Time(time.Get().Add(duration))
}

// Equal reports whether values represent the same time instant.
func (time Time) Equal(rhs Time) bool { return time.Get().Equal(rhs.Get()) }

// Get return standart time value.
func (time Time) Get() stdtime.Time { return stdtime.Time(time) }

// MarshalDynamoDBAttributeValue marshals time type for DynamoDB.
func (time Time) MarshalDynamoDBAttributeValue(
	result *dynamodb.AttributeValue,
) error {
	result.N = aws.String(strconv.FormatInt(stdtime.Time(time).Unix(), 10))
	return nil
}

// UnmarshalDynamoDBAttributeValue unmarshals time type from DynamoDB.
func (time *Time) UnmarshalDynamoDBAttributeValue(
	source *dynamodb.AttributeValue,
) error {
	seconds, err := strconv.ParseInt(*source.N, 10, 0)
	if err != nil {
		return err
	}
	*time = Time(stdtime.Unix(seconds, 0))
	return nil
}

////////////////////////////////////////////////////////////////////////////////

// DateOrTime is an object which could have date and time, or date only.
type DateOrTime struct {
	Value      Time
	IsDateOnly bool
}

// NewDateOrTime creates new DateOrTime.
func NewDateOrTime(value Time, isDateOnly bool) DateOrTime {
	return DateOrTime{
		Value:      value,
		IsDateOnly: isDateOnly,
	}
}

// Equal reports whether values represent the same time instant.
func (time DateOrTime) Equal(rhs DateOrTime) bool {
	return time.Value.Equal(rhs.Value)
}

// MarshalDynamoDBAttributeValue implements serialization for Dynamodb.
func (time DateOrTime) MarshalDynamoDBAttributeValue(
	result *dynamodb.AttributeValue,
) error {
	value := (time.Value.Get().Unix() * 10)
	if time.IsDateOnly {
		value += 1
	}
	result.N = aws.String(strconv.FormatInt(value, 10))
	return nil
}

// UnmarshalDynamoDBAttributeValue implements reading from Dynamodb.
func (time *DateOrTime) UnmarshalDynamoDBAttributeValue(
	source *dynamodb.AttributeValue,
) error {
	seconds, err := strconv.ParseInt(*source.N, 10, 0)
	if err != nil {
		return err
	}
	time.IsDateOnly = seconds%10 != 0
	if time.IsDateOnly {
		seconds -= 1
	}
	seconds /= 10
	time.Value = Time(stdtime.Unix(seconds, 0))
	return nil
}

////////////////////////////////////////////////////////////////////////////////
