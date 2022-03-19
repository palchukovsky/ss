// Copyright 2021-2022, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ss_test

import (
	"errors"
	"testing"

	"github.com/palchukovsky/ss"
	"github.com/stretchr/testify/assert"
)

func Test_SS_LogMsg(test *testing.T) {
	assert := assert.New(test)

	root := ss.NewLogMsg("root").AddErr(errors.New("root error"))
	level1 := ss.
		NewLogMsg("level 1").
		AddErr(errors.New("error 1 on level 1")).
		AddErr(errors.New("error 2 on level 1"))
	level2 := ss.NewLogMsg("level 2")

	level1.SetParent(root)
	level2.SetParent(level1)

	assert.Equal(
		"level 2: 1) error 1 on level 1; 2) error 2 on level 1; 3) root error;\n1: level 1: 1) error 1 on level 1; 2) error 2 on level 1; 3) root error;\n1: root: 1) root error;", level2.GetMessage())

}
