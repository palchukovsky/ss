// Copyright 2021-2022, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package gatewaylambda

import (
	"github.com/palchukovsky/ss"
	"github.com/palchukovsky/ss/lambda"
)

// Service implements methods of lambda service which has an output gateway.
type Service struct{ Gateway lambda.Gateway }

// NewService creates new Service instance.
func NewService() Service {
	result := Service{Gateway: lambda.NewGateway()}
	ss.S.Log().Started()
	return result
}
