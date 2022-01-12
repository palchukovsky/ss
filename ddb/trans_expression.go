// Copyright 2021-2022, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ddb

import (
	"github.com/palchukovsky/ss"
)

////////////////////////////////////////////////////////////////////////////////

type CheckedTransExpression interface {
	ss.NoCopy

	AllowConditionalCheckFail() ConditionalTransCheckFailPermission
}

////////////////////////////////////////////////////////////////////////////////

type ConditionalTransCheckFailPermission int

func (p ConditionalTransCheckFailPermission) GetIndex() int { return int(p) }

////////////////////////////////////////////////////////////////////////////////

type checkedTransExpression struct {
	ss.NoCopyImpl

	index                          int
	allowedToFailConditionalChecks []bool
}

func newCheckedTransExpression(
	expressionIndex int,
	allowedToFailConditionalChecks []bool,
) checkedTransExpression {
	return checkedTransExpression{
		index:                          expressionIndex,
		allowedToFailConditionalChecks: allowedToFailConditionalChecks,
	}
}

func (expr *checkedTransExpression) AllowConditionalCheckFail() ConditionalTransCheckFailPermission {
	expr.allowedToFailConditionalChecks[expr.index] = true
	return ConditionalTransCheckFailPermission(expr.index)
}

////////////////////////////////////////////////////////////////////////////////
