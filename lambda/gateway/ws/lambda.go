// Copyright 2021-2022, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package wsgatewaylambda

// Lambda describes lambda.
type Lambda interface{ Execute(Request) error }

////////////////////////////////////////////////////////////////////////////////
