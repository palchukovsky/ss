// Copyright 2021, the SS project owners. All rights reserved.
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
func newResult(source error) (Result, error) {
	if source == nil {
		return true, nil
	}

	{
		var awsErr awserr.Error
		if errors.As(source, &awsErr) {
			switch awsErr.Code() {
			case dynamodb.ErrCodeConditionalCheckFailedException:
				return false, nil
			}
		}
	}

	return false, source
}

////////////////////////////////////////////////////////////////////////////////
