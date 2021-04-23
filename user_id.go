// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ss

// UserIDSize is the size of UserID in bytes.
const UserIDSize = EntityIDSize

// ss.UserID is a user ID.
type UserID struct{ EntityID }

// NewUserID generates new user ID.
func NewUserID() UserID { return LoadUserID(NewEntityID()) }

// LoadUserID loads UserID from entity ID.
func LoadUserID(source EntityID) UserID { return UserID{EntityID: source} }

// LoadUserID loads UserID from entity ID.
func ParseUserID(source string) (UserID, error) {
	entityID, err := ParseEntityID(source)
	if err != nil {
		return UserID{}, err
	}
	return LoadUserID(entityID), nil
}
