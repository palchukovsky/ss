// Copyright 2021-2022, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ss

import (
	"encoding/base64"
	"fmt"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/google/uuid"
)

////////////////////////////////////////////////////////////////////////////////

// EntityIDSize is the size of EntityID in bytes.
const EntityIDSize = 16

// EntityID is an abstract project entity ID.
type EntityID uuid.UUID

// NewEntityID generates new entity ID.
func NewEntityID() EntityID { return EntityID(uuid.New()) }

// ParseEntityID parses ID in string.
func ParseEntityID(source string) (EntityID, error) {
	bin, err := base64.RawStdEncoding.DecodeString(source)
	if err != nil {
		return EntityID{}, fmt.Errorf(`failed to decode entity ID: "%w"`, err)
	}
	result := EntityID{}
	if err := result.UnmarshalBinary(bin); err != nil {
		return EntityID{}, fmt.Errorf(`failed to unmarshal entity ID: "%w"`, err)
	}
	return result, nil
}

// String returns ID as string.
func (id EntityID) String() string {
	return base64.RawStdEncoding.EncodeToString(id[:])
}

// MarshalText implements encoding.TextMarshaler.
func (id EntityID) MarshalText() ([]byte, error) {
	return []byte(id.String()), nil
}

// UnmarshalText implements encoding.TextUnmarshaler.
func (id *EntityID) UnmarshalText(data []byte) error {
	var err error
	*id, err = ParseEntityID(string(data))
	return err
}

// MarshalBinary implements encoding.BinaryMarshaler.
func (id EntityID) MarshalBinary() ([]byte, error) {
	return uuid.UUID(id).MarshalBinary()
}

// UnmarshalBinary implements encoding.BinaryUnmarshaler.
func (id *EntityID) UnmarshalBinary(data []byte) error {
	uuid := uuid.UUID{}
	if err := uuid.UnmarshalBinary(data); err != nil {
		return err
	}
	*id = EntityID(uuid)
	return nil
}

// MarshalDynamoDBAttributeValue implements serialization for Dynamodb.
func (id EntityID) MarshalDynamoDBAttributeValue(
	result *dynamodb.AttributeValue,
) error {
	var err error
	result.B, err = id.MarshalBinary()
	return err
}

// UnmarshalDynamoDBAttributeValue implements reading from Dynamodb.
func (id *EntityID) UnmarshalDynamoDBAttributeValue(
	source *dynamodb.AttributeValue,
) error {
	if source.S != nil {
		// Entity ID could be represent as text.
		return id.UnmarshalText([]byte(*source.S))
	}
	return id.UnmarshalBinary(source.B)
}

////////////////////////////////////////////////////////////////////////////////
