// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ss

// S is the standalone service instance.
var S Service

// Set sets service without any initialization and checks.
func Set(s Service) { S = s }
