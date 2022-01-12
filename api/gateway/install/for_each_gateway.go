// Copyright 2021-2022, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package apigatewayinstall

import (
	"fmt"

	"github.com/palchukovsky/ss"
	gatewayinstall "github.com/palchukovsky/ss/gateway/install"
)

// ForEachGateway calls the callback for each gateway logs it.
func ForEachGateway(
	installer Installer,
	callback func(gatewayinstall.Gateway) error,
	log ss.Log,
) error {
	log.Debug(ss.NewLogMsg("processing each gateway..."))

	gateways := installer.NewGateways(log)
	if id := ss.S.Config().AWS.Gateway.App.ID; id != "" {
		gateways = append(gateways, gatewayinstall.NewGateway(id, "app", log))
	}

	for _, gateway := range gateways {
		gateway.Log().Debug(ss.NewLogMsg("Processing..."))
		if err := callback(gateway); err != nil {
			return fmt.Errorf(`failed to process %q: "%w"`, gateway.GetName(), err)
		}
		gateway.Log().Info(ss.NewLogMsg("processed"))
	}

	log.Debug(ss.NewLogMsg("processing each gateway successfully completed"))
	return nil
}
