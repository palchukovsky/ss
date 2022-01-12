// Copyright 2021-2022, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ddb

import (
	"errors"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/dynamodb"
)

type Result bool

func (result Result) IsSuccess() bool { return bool(result) }

////////////////////////////////////////////////////////////////////////////////

// newResult creates a result object if possible.
// In case of an unexpected error, returns clarified error.
func newResult(err error, isConditionalCheckFailAllowed bool) (Result, error) {
	if err == nil {
		return true, nil
	}

	if isConditionalCheckFailAllowed {
		var awsErr awserr.Error
		if errors.As(err, &awsErr) {
			switch awsErr.Code() {
			case dynamodb.ErrCodeConditionalCheckFailedException:
				return false, nil
			}
		}
	}

	return false, err
}

////////////////////////////////////////////////////////////////////////////////
