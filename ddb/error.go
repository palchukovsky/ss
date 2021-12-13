// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ddb

import (
	"errors"
	"strings"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

// isConditionalCheckError retruns true if error is the error
// at conditions check.
func isConditionalCheckError(source error) bool {
	var awsErr awserr.Error
	if !errors.As(source, &awsErr) {
		return false
	}
	switch awsErr.Code() {
	case dynamodb.ErrCodeConditionalCheckFailedException,
		dynamodb.ErrCodeTransactionCanceledException:
		{
			return true
		}
	}
	return false
}

// parseErrorConditionalCheckFailed parses error to check
// what condition was failed. Returns nil if it's not a error "condition failed"
// or if failed conditions outside provided range.
func parseErrorConditionalCheckFailed(
	source error,
	conditionFromIndex int,
	conditionsNumber int,
) []bool {

	var awsErr awserr.Error
	if !errors.As(source, &awsErr) {
		return nil
	}
	if awsErr.Code() != dynamodb.ErrCodeTransactionCanceledException {
		return nil
	}

	message := awsErr.Message()
	begin := strings.LastIndex(message, "[")
	end := strings.LastIndex(message, "]")
	if begin >= end {
		return nil
	}

	conditionResults := strings.Split(message[begin+1:end], ",")
	var conditionTotalLen int
	if conditionsNumber != 0 {
		conditionTotalLen = conditionFromIndex + conditionsNumber
		if len(conditionResults) < conditionTotalLen {
			return nil
		}
	} else {
		conditionTotalLen = len(conditionResults)
	}

	result := make([]bool, conditionsNumber)
	for i := 0; i < len(conditionResults); i++ {
		var isOk bool
		switch strings.TrimSpace(conditionResults[i]) {
		case "None":
			isOk = true
		case "ConditionalCheckFailed":
			isOk = false
		default:
			return nil
		}
		if i < conditionFromIndex || i >= conditionTotalLen {
			if !isOk {
				return nil
			}
			continue
		}
		result[i-conditionFromIndex] = isOk
	}
	return result
}
