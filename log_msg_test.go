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

	root1 := ss.NewLogMsg("root 1").AddErr(errors.New("root error 1"))
	root2 := ss.NewLogMsg("root 2").AddErr(errors.New("root error 2"))

	level1_1 := ss.
		NewLogMsg("level 1.1").
		AddErr(errors.New("error 1 on level 1.1")).
		AddErr(errors.New("error 2 on level 1.1"))
	level1_2 := ss.
		NewLogMsg("level 1.2").
		AddErr(errors.New("error 1 on level 1.2")).
		AddErr(errors.New("error 2 on level 1.2"))

	level2 := ss.NewLogMsg("level 2")

	level1_1.AddParent(root1)
	level1_1.AddParent(root2)
	level2.AddParent(level1_1)
	level2.AddParent(level1_2)

	/*
		level 2: 1) error 1 on level 1.1; 2) error 2 on level 1.1; 3) root error 1; 4) root error 2; 5) error 1 on level 1.2; 6) error 2 on level 1.2;
		1: level 1.1: 1) error 1 on level 1.1; 2) error 2 on level 1.1; 3) root error 1; 4) root error 2;
		1.1: root 1: 1) root error 1;
		1.2: root 2: 1) root error 2;
		2: level 1.2: 1) error 1 on level 1.2; 2) error 2 on level 1.2;
	*/
	assert.Equal(
		"level 2: 1) error 1 on level 1.1; 2) error 2 on level 1.1; 3) root error 1; 4) root error 2; 5) error 1 on level 1.2; 6) error 2 on level 1.2;\n1: level 1.1: 1) error 1 on level 1.1; 2) error 2 on level 1.1; 3) root error 1; 4) root error 2;\n1.1: root 1: 1) root error 1;\n1.2: root 2: 1) root error 2;\n2: level 1.2: 1) error 1 on level 1.2; 2) error 2 on level 1.2;",
		level2.GetMessage())

}
