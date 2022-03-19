// Copyright 2021-2022, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ss_test

import (
	"testing"

	"github.com/palchukovsky/ss"
	"github.com/stretchr/testify/assert"
)

func Test_SS_LogMsg(test *testing.T) {
	assert := assert.New(test)

	root := ss.NewLogMsg("root")
	level1 := ss.NewLogMsg("level 1")
	level2 := ss.NewLogMsg("level 2")

	result := root.MergeWithLowLevelMsg(
		level1.MergeWithLowLevelMsg(
			level2))

	assert.Equal("root\n1: level 1\n2: level 2", result.GetMessage())

}
