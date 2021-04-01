// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ddb

////////////////////////////////////////////////////////////////////////////////

// SubscriptionVersion is a version of subscription.
type SubscriptionVersion uint

// NewSubscriptionVersion creates new subscription version instance.
func NewSubscriptionVersion(source uint) SubscriptionVersion {
	return SubscriptionVersion(source)
}

////////////////////////////////////////////////////////////////////////////////
