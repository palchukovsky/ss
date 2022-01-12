// Copyright 2021-2022, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ddb

import "github.com/palchukovsky/ss"

////////////////////////////////////////////////////////////////////////////////

type CheckedExpression interface {
	ss.NoCopy

	AllowConditionalCheckFail()
}

////////////////////////////////////////////////////////////////////////////////

type checkedExpression struct {
	ss.NoCopyImpl

	isConditionalCheckFailAllowed bool
}

func newCheckedExpression() checkedExpression { return checkedExpression{} }

func (expr *checkedExpression) AllowConditionalCheckFail() {
	expr.isConditionalCheckFailAllowed = true
}

////////////////////////////////////////////////////////////////////////////////
