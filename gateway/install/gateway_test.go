// Copyright 2021-2022, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package gatewayinstall_test

import (
	"testing"

	gatewayinstall "github.com/palchukovsky/ss/gateway/install"
	"github.com/stretchr/testify/assert"
)

func Test_GateWay_Install_GatewayCommandName(test *testing.T) {
	assert := assert.New(test)

	name := gatewayinstall.NewGatewayCommandName(
		"asd/123/qwerty/123/zxCv asdf",
		"asd/123/")
	assert.Equal("Qwerty123ZxCvAsdf", name)
}
