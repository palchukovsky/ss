// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ss

// ConnectionID is a connection ID.
type ConnectionID string

func (id ConnectionID) MarshalLogMsg(destination map[string]interface{}) {
	destination[logMsgNodeConnection] = id
}
