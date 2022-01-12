// Copyright 2021-2022, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package apigatewayinstall

import (
	"github.com/palchukovsky/ss"
	gatewayinstall "github.com/palchukovsky/ss/gateway/install"
)

// Installer describes the API getway installing interface.
type Installer interface {
	NewGateways(ss.Log) []gatewayinstall.Gateway
}
