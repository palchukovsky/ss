// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package creategatewaylambda

import (
	"github.com/palchukovsky/ss"
	api "github.com/palchukovsky/ss/api/gateway/install"
	install "github.com/palchukovsky/ss/gateway/install"
)

func Init(
	initService func(projectPackage string, params ss.ServiceParams),
) {
	initService("install", ss.ServiceParams{})
}

func Run(installer api.Installer) {
	log := ss.S.Log()
	defer func() { log.CheckExit(recover()) }()
	log.Started()

	client := install.NewClient()

	err := api.ForEachGateway(
		installer,
		func(gateway install.Gateway) error { return gateway.Create(client) },
		log)
	if err != nil {
		log.Panic(ss.NewLogMsg(`failed to create gateway`).AddErr(err))
	}

}
