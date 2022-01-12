// Copyright 2021-2022, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ss

// BoolPtr converts bool value into pointer on bool.
func BoolPtr(value bool) *bool { return &value }

// BoolPtrIfSet converts bool value into pointer on bool.
func BoolPtrIfSet(value bool) *bool {
	if !value {
		return nil
	}
	return BoolPtr(true)
}

// IsBoolSet returns true if bool set and true.
func IsBoolSet(value *bool) bool { return value != nil && *value }
