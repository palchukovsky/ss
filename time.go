// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ss

import (
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"time"
	stdtime "time"

	aws "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

////////////////////////////////////////////////////////////////////////////////

type Time struct{ value stdtime.Time }

func NewTime(source stdtime.Time) Time { return Time{value: source} }

func NewTimeFromDate(source Date) Time {
	return NewTime(
		time.Date(
			int(source.year), source.month, int(source.day), 0, 0, 0, 0, time.UTC))
}

// Now return current platform time in correct time zone.
func Now() Time                        { return NewTime(stdtime.Now().UTC()) }
func (time Time) Get() stdtime.Time    { return time.value }
func (time Time) Equal(rhs Time) bool  { return time.value.Equal(rhs.value) }
func (time Time) Before(rhs Time) bool { return time.value.Before(rhs.value) }
func (time Time) After(rhs Time) bool  { return time.value.After(rhs.value) }
func (time Time) Unix() int64          { return time.value.Unix() }
func (time Time) UTC() Time            { return NewTime(time.value.UTC()) }

func (time Time) Add(duration time.Duration) Time {
	return NewTime(time.value.Add(duration))
}

func (time Time) AddDate(years int, months int, days int) Time {
	return NewTime(time.value.AddDate(years, months, days))
}

func (time Time) String() string {
	year, month, day := time.value.Date()
	return fmt.Sprintf(
		"%d%02d%02d %02d%02d%02d",
		year,
		month,
		day,
		time.value.Hour(),
		time.value.Minute(),
		time.value.Second())
}

func (time Time) MarshalText() ([]byte, error) {
	return []byte(time.String()), nil
}

func (time *Time) UnmarshalText(source []byte) error {
	if len(source) == 0 {
		*time = Time{}
		return nil
	}

	match := regexp.
		MustCompile(
			`^(\d{4})(0[1-9]|1[0-2])(0[1-9]|[1-2][\d]|3[0-1])( ([0-1][\d]|2[0-3])([0-5][\d])([0-5][\d]))?$`).
		FindStringSubmatch(string(source))
	if len(match) == 0 {
		return fmt.Errorf(`failed to parse time %q`, string(source))
	}

	var year int
	var month int
	var day int
	var hour int
	var minute int
	var seconds int
	var err error
	if year, err = strconv.Atoi(match[1]); err != nil {
		return fmt.Errorf(`failed to parse year %q`, string(source))
	}
	if month, err = strconv.Atoi(match[2]); err != nil {
		return fmt.Errorf(`failed to parse month %q`, string(source))
	}
	if day, err = strconv.Atoi(match[3]); err != nil {
		return fmt.Errorf(`failed to parse day %q`, string(source))
	}
	if hour, err = strconv.Atoi(match[5]); err != nil {
		return fmt.Errorf(`failed to parse hours %q`, string(source))
	}
	if minute, err = strconv.Atoi(match[6]); err != nil {
		return fmt.Errorf(`failed to parse minutes %q`, string(source))
	}
	if seconds, err = strconv.Atoi(match[7]); err != nil {
		return fmt.Errorf(`failed to parse seconds %q`, string(source))
	}

	*time = NewTime(
		stdtime.Date(
			year, stdtime.Month(month), day, hour, minute, seconds, 0, stdtime.UTC))

	return nil
}

func (time Time) MarshalDynamoDBAttributeValue(
	result *dynamodb.AttributeValue,
) error {
	result.N = aws.String(strconv.FormatInt(time.value.Unix(), 10))
	return nil
}

// UnmarshalDynamoDBAttributeValue implements reading from Dynamodb.
func (time *Time) UnmarshalDynamoDBAttributeValue(
	source *dynamodb.AttributeValue,
) error {
	if source.N == nil {
		return errors.New("DynamoDB value is not number")
	}
	seconds, err := strconv.ParseInt(*source.N, 10, 0)
	if err != nil {
		return err
	}
	time.value = stdtime.Unix(seconds, 0)
	return nil
}

////////////////////////////////////////////////////////////////////////////////

// DateOrTime is an object which could have date and time, or date only.
type DateOrTime struct {
	Value      Time
	IsDateOnly bool
	// After each new field check the method IsEqual.
}

// NewDateOrTime creates new DateOrTime.
func NewDateOrTime(value Time, isDateOnly bool) DateOrTime {
	return DateOrTime{
		Value:      value,
		IsDateOnly: isDateOnly,
	}
}

// IsEqual reports whether values represent the same time instant.
func (time DateOrTime) IsEqual(rhs DateOrTime) bool {
	return time.IsDateOnly == rhs.IsDateOnly && time.Value.Equal(rhs.Value)
}

func (time DateOrTime) String() string {
	if time.IsDateOnly {
		year, month, day := time.Value.Get().Date()
		return fmt.Sprintf("%d%02d%02d", year, month, day)
	}
	return time.Value.String()
}

func (time *DateOrTime) MarshalText() ([]byte, error) {
	return []byte(time.String()), nil
}

func (time *DateOrTime) UnmarshalText(source []byte) error {

	if len(source) > 4+1+2+1+2 /* 2009-01-07 */ {
		if err := time.Value.UnmarshalText(source); err != nil {
			return err
		}
		time.IsDateOnly = false
		return nil
	}

	if len(source) == 0 {
		time.Value = Time{}
		return nil
	}

	match := regexp.
		MustCompile(`^(\d{4})(0[1-9]|1[0-2])(0[1-9]|[1-2][\d]|3[0-1])?$`).
		FindStringSubmatch(string(source))
	if len(match) == 0 {
		return fmt.Errorf(`failed to parse event time %q`, string(source))
	}
	var year int
	var month int
	var day int
	var err error
	if year, err = strconv.Atoi(match[1]); err != nil {
		return fmt.Errorf(`failed to parse year in event time %q`, string(source))
	}
	if month, err = strconv.Atoi(match[2]); err != nil {
		return fmt.Errorf(`failed to parse year in event time %q`, string(source))
	}
	if day, err = strconv.Atoi(match[3]); err != nil {
		return fmt.Errorf(`failed to parse day in event time %q`, string(source))
	}

	time.Value = NewTime(
		stdtime.Date(year, stdtime.Month(month), day, 0, 0, 0, 0, stdtime.UTC))
	time.IsDateOnly = true

	return nil
}

// MarshalDynamoDBAttributeValue implements serialization for Dynamodb.
func (time DateOrTime) MarshalDynamoDBAttributeValue(
	result *dynamodb.AttributeValue,
) error {
	value := (time.Value.Get().Unix() * 10)
	if !time.IsDateOnly {
		// The same time at the day start will be placed in the order:
		// 1) only date, as value will be 100
		// 2) date + time, as value will be 101
		// So, to filter "only by date" query should use "= 100",
		// for time ">= 101".
		value += 1
	}
	result.N = aws.String(strconv.FormatInt(value, 10))
	return nil
}

// UnmarshalDynamoDBAttributeValue implements reading from Dynamodb.
func (time *DateOrTime) UnmarshalDynamoDBAttributeValue(
	source *dynamodb.AttributeValue,
) error {
	if source.N == nil {
		return errors.New("DynamoDB value is not number")
	}
	seconds, err := strconv.ParseInt(*source.N, 10, 0)
	if err != nil {
		return err
	}
	time.IsDateOnly = seconds%10 == 0
	if !time.IsDateOnly {
		seconds -= 1
	}
	seconds /= 10
	time.Value = NewTime(stdtime.Unix(seconds, 0))
	return nil
}

////////////////////////////////////////////////////////////////////////////////

type Date struct {
	year  uint
	month time.Month
	day   uint
}

func NewDate(year uint, month time.Month, day uint) Date {
	return Date{
		year:  year,
		month: month,
		day:   day,
	}
}

func NewDateFromTime(source time.Time) Date {
	return NewDate(uint(source.Year()), source.Month(), uint(source.Day()))
}

func (date Date) Year() uint        { return date.year }
func (date Date) Month() time.Month { return date.month }
func (date Date) Day() uint         { return date.day }

func (date Date) After(rhs Date) bool { return date.Number() > rhs.Number() }

func (date Date) Number() uint {
	return (date.year * 10000) + (uint(date.month) * 100) + date.day
}

func (date Date) String() string {
	return fmt.Sprintf("%d%02d%02d", date.year, date.month, date.day)
}

func (date Date) MarshalText() ([]byte, error) {
	return []byte(date.String()), nil
}

func (date *Date) UnmarshalText(source []byte) error {
	if len(source) == 0 {
		*date = Date{}
		return nil
	}

	match := regexp.
		MustCompile(`^(\d{4})(0[1-9]|1[0-2])(0[1-9]|[1-2][\d]|3[0-1])$`).
		FindStringSubmatch(string(source))
	if len(match) == 0 {
		return fmt.Errorf(`failed to parse date %q`, string(source))
	}

	var year uint64
	var month uint64
	var day uint64
	var err error
	if year, err = strconv.ParseUint(match[1], 10, 0); err != nil {
		return fmt.Errorf(`failed to parse year %q`, string(source))
	}
	if month, err = strconv.ParseUint(match[2], 10, 0); err != nil {
		return fmt.Errorf(`failed to parse month %q`, string(source))
	}
	if day, err = strconv.ParseUint(match[3], 10, 0); err != nil {
		return fmt.Errorf(`failed to parse day %q`, string(source))
	}

	*date = NewDate(uint(year), stdtime.Month(month), uint(day))

	return nil
}

func (date Date) MarshalDynamoDBAttributeValue(
	result *dynamodb.AttributeValue,
) error {
	result.N = aws.String(strconv.FormatUint(uint64(date.Number()), 10))
	return nil
}

func (date *Date) UnmarshalDynamoDBAttributeValue(
	source *dynamodb.AttributeValue,
) error {
	if source.N == nil {
		return errors.New("DynamoDB value is not number")
	}
	number, err := strconv.ParseUint(*source.N, 10, 0)
	if err != nil {
		return err
	}
	*date = NewDate(
		uint(number/10000),
		stdtime.Month((number/100)%100),
		uint(number%10))
	return nil
}

////////////////////////////////////////////////////////////////////////////////
