// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package main

import (
	"flag"

	"github.com/palchukovsky/ss"
	lambda "github.com/palchukovsky/ss/api/gateway/install/lambda/create"
)

////////////////////////////////////////////////////////////////////////////////

var (
	accountID       = flag.String("account", "", "AWS account ID")
	region          = flag.String("region", "", "AWS region")
	accessKeyID     = flag.String("key", "", "AWS access key ID")
	accessKeySecret = flag.String("secret", "", "AWS access key secret")
	restGatewayID   = flag.String("secret", "",
		"AWS API gateway ID for REST endpoint")
)

////////////////////////////////////////////////////////////////////////////////

func init() {
	lambda.Init(
		func(projectPackage string, params ss.ServiceParams) {
			ss.Set(newService(projectPackage, ss.ServiceConfig{
				AWS: ss.AWSConfig{
					AccountID: *accountID,
					Region:    *region,
					AccessKey: ss.NewAWSAccessKey(*accessKeyID, *accessKeySecret),
				},
			}))
		})
}

func main() { lambda.Run(newInstaller(*restGatewayID)) }

////////////////////////////////////////////////////////////////////////////////
