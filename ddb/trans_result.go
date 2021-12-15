// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ddb

import (
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/palchukovsky/ss"
)

type TransResult interface {
	IsSuccess() bool
	ParseConditions() ConditionalCheckResult

	MarshalLogMsg(destination map[string]interface{})
}

type ConditionalCheckResult interface {
	IsPassed(conditions ...WriteTransExpression) bool
	IsFailedOnly(allowedToFail ...WriteTransExpression) bool
}

////////////////////////////////////////////////////////////////////////////////

// newTransResult creates a result object if possible.
// In case of an unexpected error, returns clarified error.
func newTransResult(source error) (TransResult, error) {
	if source == nil {
		return newSuccessfulTransResult(), nil
	}

	{
		var awsErr awserr.Error
		if errors.As(source, &awsErr) {
			switch awsErr.Code() {
			case dynamodb.ErrCodeConditionalCheckFailedException,
				dynamodb.ErrCodeTransactionCanceledException:
				{
					return newConditionalTransCheckFail(awsErr), nil
				}
			}
		}
	}

	return nil, source
}

////////////////////////////////////////////////////////////////////////////////

type successfulTransResult struct{}

func newSuccessfulTransResult() successfulTransResult {
	return successfulTransResult{}
}
func (successfulTransResult) IsSuccess() bool { return true }
func (successfulTransResult) ParseConditions() ConditionalCheckResult {
	return successfullyTestedConditions{}
}
func (successfulTransResult) MarshalLogMsg(destination map[string]interface{}) {
	ss.MarshalLogMsgAttrDump(nil, destination)
}

type successfullyTestedConditions struct{}

func (successfullyTestedConditions) IsPassed(...WriteTransExpression) bool {
	return true
}
func (successfullyTestedConditions) IsFailedOnly(...WriteTransExpression) bool {
	return true
}

////////////////////////////////////////////////////////////////////////////////

type conditionalTransCheckFail struct{ source awserr.Error }

func newConditionalTransCheckFail(
	source awserr.Error,
) conditionalTransCheckFail {
	return conditionalTransCheckFail{source: source}
}

func (conditionalTransCheckFail) IsSuccess() bool { return false }

func (fail conditionalTransCheckFail) ParseConditions() ConditionalCheckResult {

	message := fail.source.Message()
	begin := strings.LastIndex(message, "[")
	end := strings.LastIndex(message, "]")
	if begin >= end {
		ss.S.Log().Panic(
			ss.
				NewLogMsg("failed to parse conditional check fail message").
				AddErr(fail.source).
				AddDump(message))
	}

	conditions := strings.Split(message[begin+1:end], ",")

	result := conditionalCheckFails{
		fail:  fail,
		flags: make([]bool, len(conditions)),
	}

	for i, condition := range conditions {
		switch strings.TrimSpace(condition) {
		case "None":
			result.flags[i] = true
		case "ConditionalCheckFailed":
			result.flags[i] = false
		default:
			ss.S.Log().Panic(
				ss.
					NewLogMsg(
						"unknown conditional check fail message status %q",
						condition).
					AddErr(fail.source).
					AddDump(message))
		}
	}

	return result
}

func (fail conditionalTransCheckFail) MarshalLogMsg(destination map[string]interface{}) {
	ss.MarshalLogMsgAttrDump(fail.source, destination)
}

type conditionalCheckFails struct {
	fail  conditionalTransCheckFail
	flags []bool
}

func (checks conditionalCheckFails) IsPassed(
	conditions ...WriteTransExpression,
) bool {
	for _, condition := range conditions {
		if !checks.flags[condition.GetIndex()] {
			return false
		}
	}
	return true
}

func (check conditionalCheckFails) IsFailedOnly(
	allowedToFail ...WriteTransExpression,
) bool {
	for i, flag := range check.flags {
		if flag {
			continue
		}
		failIsAllowed := false
		for _, condition := range allowedToFail {
			if condition.GetIndex() == i {
				failIsAllowed = true
				break
			}
		}
		if failIsAllowed {
			return false
		}
	}
	return true
}

////////////////////////////////////////////////////////////////////////////////
