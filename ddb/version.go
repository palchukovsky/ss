// Copyright 2021-2022, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ddb

////////////////////////////////////////////////////////////////////////////////

// MembershipVersion is a version of membership.
type MembershipVersion uint

// NewMembershipVersion creates new membership version instance.
func NewMembershipVersion(source uint) MembershipVersion {
	return MembershipVersion(source)
}

////////////////////////////////////////////////////////////////////////////////
