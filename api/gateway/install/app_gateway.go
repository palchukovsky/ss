// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package apigatewayinstall

import (
	"github.com/palchukovsky/ss"
	gatewayinstall "github.com/palchukovsky/ss/gateway/install"
)

func newAppGateway(id string, log ss.ServiceLogStream) gatewayinstall.Gateway {
	return gatewayinstall.NewGateway(
		id,
		"app",
		gatewayinstall.NewGatewayCommadsReader(gatewayinstall.NewWSCommand),
		log)
}
