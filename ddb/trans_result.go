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
	IsPassed(conditions ...ConditionalTransCheckFailPermission) bool
}

////////////////////////////////////////////////////////////////////////////////

// newTransResult creates a result object if possible.
// In case of an unexpected error, returns clarified error.
func newTransResult(err error, trans WriteTrans) (TransResult, error) {

	if err == nil {
		return newSuccessfulTransResult(trans), nil
	}

	{
		var awsErr awserr.Error
		if errors.As(err, &awsErr) {
			switch awsErr.Code() {
			case dynamodb.ErrCodeConditionalCheckFailedException,
				dynamodb.ErrCodeTransactionCanceledException:
				{
					result, ok := newConditionalTransCheckFail(awsErr, trans)
					if !ok {
						return nil, err
					}
					return result, nil
				}
			}
		}
	}

	return nil, err
}

////////////////////////////////////////////////////////////////////////////////

type successfulTransResult struct{ trans WriteTrans }

func newSuccessfulTransResult(trans WriteTrans) successfulTransResult {
	return successfulTransResult{trans: trans}
}
func (successfulTransResult) IsSuccess() bool { return true }
func (successfulTransResult) ParseConditions() ConditionalCheckResult {
	return successfullyTestedConditions{}
}
func (success successfulTransResult) MarshalLogMsg(
	destination map[string]interface{},
) {
	ss.MarshalLogMsgAttrDump(success.trans, destination)
}

type successfullyTestedConditions struct{}

func (successfullyTestedConditions) IsPassed(
	...ConditionalTransCheckFailPermission,
) bool {
	return true
}

////////////////////////////////////////////////////////////////////////////////

type conditionalTransCheckFail struct {
	err                    awserr.Error
	trans                  WriteTrans
	conditionalCheckResult ConditionalCheckResult
}

func newConditionalTransCheckFail(
	err awserr.Error,
	trans WriteTrans,
) (
	conditionalTransCheckFail,
	bool,
) {
	result := conditionalTransCheckFail{err: err}
	allowedToFailConditionalCheck := trans.getAllowedToFailConditionalChecks()
	return result,
		len(allowedToFailConditionalCheck) == 0 ||
			result.parseConditions(allowedToFailConditionalCheck)
}

func (conditionalTransCheckFail) IsSuccess() bool { return false }

func (fail conditionalTransCheckFail) ParseConditions() ConditionalCheckResult {
	if fail.conditionalCheckResult == nil {
		fail.parseConditions(nil)
	}
	return fail.conditionalCheckResult
}

func (fail *conditionalTransCheckFail) parseConditions(
	allowedToFailConditionalChecks map[int]struct{},
) bool {

	message := fail.err.Message()
	begin := strings.LastIndex(message, "[")
	end := strings.LastIndex(message, "]")
	if begin >= end {
		ss.S.Log().Panic(
			ss.
				NewLogMsg("failed to parse conditional check fail message").
				Add(fail).
				AddDump(message))
	}

	conditions := strings.Split(message[begin+1:end], ",")

	result := conditionalCheckFails{flags: make([]bool, len(conditions))}

	for i, condition := range conditions {
		switch strings.TrimSpace(condition) {
		case "None":
			result.flags[i] = true
		case "ConditionalCheckFailed":
			if allowedToFailConditionalChecks != nil {
				// nil-check is required as if nil - means "do not check"
				if _, has := allowedToFailConditionalChecks[i]; !has {
					return false
				}
			}
			result.flags[i] = false
		default:
			ss.S.Log().Panic(
				ss.
					NewLogMsg(
						"unknown conditional check fail message status %q",
						condition).
					Add(fail).
					AddDump(message))
		}

	}

	fail.conditionalCheckResult = result

	return true
}

func (fail conditionalTransCheckFail) MarshalLogMsg(
	destination map[string]interface{},
) {
	ss.MarshalLogMsgAttrDump(fail.err, destination)
	ss.MarshalLogMsgAttrDump(fail.trans, destination)
}

type conditionalCheckFails struct{ flags []bool }

func (checks conditionalCheckFails) IsPassed(
	conditions ...ConditionalTransCheckFailPermission,
) bool {
	for _, condition := range conditions {
		if !checks.flags[condition.GetIndex()] {
			return false
		}
	}
	return true
}

////////////////////////////////////////////////////////////////////////////////
