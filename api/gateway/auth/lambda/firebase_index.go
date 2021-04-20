// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package authlambda

import (
	ss "github.com/palchukovsky/ss"
	ssdb "github.com/palchukovsky/ss/db"
)

type FirebaseIndex struct {
	ssdb.UserExternalFirabaseIDIndex
	ID          ss.UserID `json:"id"`
	Name        string    `json:"name"`
	Email       string    `json:"email,omitempty"`
	PhoneNumber string    `json:"phone,omitempty"`
	PhotoURL    string    `json:"photoUrl,omitempty"`
}

func NewFirebaseIndex(source ssdb.User) FirebaseIndex {
	return FirebaseIndex{
		ID:          source.ID,
		Name:        source.Name,
		Email:       source.Email,
		PhoneNumber: source.PhoneNumber,
		PhotoURL:    source.PhotoURL,
	}
}

func (r *FirebaseIndex) Clear() { *r = FirebaseIndex{} }
