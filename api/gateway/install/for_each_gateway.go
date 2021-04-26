// Copyright 2021, the SS project owners. All rights reserved.
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
	log ss.ServiceLog,
) error {
	log.Debug("Processing each gateway...")

	gateways := append(
		installer.NewGateways(log),
		newAuthGateway(log),
		newAppGateway(log))

	for _, gateway := range gateways {
		gateway.Log().Debug("Processing...")
		if err := callback(gateway); err != nil {
			return fmt.Errorf(`failed to process %q: "%w"`, gateway.GetName(), err)
		}
		gateway.Log().Info("Processed.")
	}

	log.Debug("Processing each gateway successfully completed.")
	return nil
}
