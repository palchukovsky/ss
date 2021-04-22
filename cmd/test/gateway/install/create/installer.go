// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package main

import (
	"github.com/palchukovsky/ss"
	gatewayinstall "github.com/palchukovsky/ss/gateway/install"
)

type installer struct{ restGatewayID string }

func newInstaller(restGatewayID string) installer {
	return installer{restGatewayID: restGatewayID}
}

func (installer installer) NewGateways(
	log ss.ServiceLog,
) []gatewayinstall.Gateway {
	return []gatewayinstall.Gateway{
		gatewayinstall.NewGateway(
			installer.restGatewayID,
			"RESTTest",
			newRESTGatewayCommadsReader(),
			log),
	}
}

////////////////////////////////////////////////////////////////////////////////

func newRESTGatewayCommadsReader() gatewayinstall.GatewayCommadsReader {
	return gatewayRESTCommadsReader{}
}

type gatewayRESTCommadsReader struct{}

func (gatewayRESTCommadsReader) Read(
	name string,
	log ss.ServiceLog,
) ([]gatewayinstall.Command, error) {
	cmd, err := gatewayinstall.NewRESTCommand(
		"TestCmd",
		"./",
		log)
	if err != nil {
		return nil, err
	}
	return []gatewayinstall.Command{cmd}, nil
}

////////////////////////////////////////////////////////////////////////////////
