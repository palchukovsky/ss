// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package main

import (
	"github.com/palchukovsky/ss"
	gatewayinstall "github.com/palchukovsky/ss/gateway/install"
)

type installer struct{ wsGatewayID string }

func newInstaller(wsGatewayID string) installer {
	return installer{wsGatewayID: wsGatewayID}
}

func (installer installer) NewGateways(
	log ss.ServiceLog,
) []gatewayinstall.Gateway {
	return []gatewayinstall.Gateway{
		gatewayinstall.NewGateway(
			installer.wsGatewayID,
			"WSTest",
			newGatewayCommadsReader(),
			log),
	}
}

////////////////////////////////////////////////////////////////////////////////

func newGatewayCommadsReader() gatewayinstall.GatewayCommadsReader {
	return gatewayCommadsReader{}
}

type gatewayCommadsReader struct{}

func (gatewayCommadsReader) Read(
	name string,
	log ss.ServiceLogStream,
) ([]gatewayinstall.Command, error) {
	cmd, err := gatewayinstall.NewWSCommand(
		"TestCmd",
		"./",
		log)
	if err != nil {
		return nil, err
	}
	return []gatewayinstall.Command{cmd}, nil
}

////////////////////////////////////////////////////////////////////////////////
