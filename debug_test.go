// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ss_test

import (
	"testing"

	"github.com/palchukovsky/ss"
	"github.com/stretchr/testify/assert"
)

type TestType struct {
}

func Test_SS_Debug_GetTypeName(test *testing.T) {
	assert := assert.New(test)

	assert.Equal("github.com/palchukovsky/ss_test/TestType",
		ss.GetTypeName(TestType{}))
	assert.Equal("*github.com/palchukovsky/ss_test/TestType",
		ss.GetTypeName(&TestType{}))

	intVal := 10
	assert.Equal("int", ss.GetTypeName(intVal))
	assert.Equal("*int", ss.GetTypeName(&intVal))

	listVal := []int{10}
	assert.Equal("", ss.GetTypeName(listVal))
	assert.Equal("*", ss.GetTypeName(&listVal))
}
