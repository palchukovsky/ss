// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package apidbevent_test

import (
	"encoding/json"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	"github.com/palchukovsky/ss"
	apidbevent "github.com/palchukovsky/ss/api/dbevent"
	"github.com/palchukovsky/ss/ddb"
	"github.com/stretchr/testify/assert"
)

type dbeventUnmarshalTestRecord struct {
	ID        ss.EntityID  `json:"id"`
	Spot      ss.EntityID  `json:"spot"`
	Event     ss.EntityID  `json:"event"`
	Partition [1 + 16]byte `json:"partition"`
	User      ss.UserID    `json:"user"`
	Rights    struct {
		Own *struct {
			Base uint `json:"base"`
		} `json:"own,omitempty"`
		Spot struct {
			Base uint `json:"base"`
		} `json:"spot"`
	} `json:"rights"`
	Version  ddb.SubscriptionVersion `json:"ver"`
	Snapshot struct {
		StartTime *ddb.DateOrTime `json:"start,omitempty"`
		Location  *struct {
			LatLng [2]float32 `json:"latLng"`
			Name   []string   `json:"name,omitempty"`
		} `json:"loc,omitempty"`
		Description    string `json:"desc,omitempty"`
		Link           string `json:"link,omitempty"`
		SpotVisibility uint   `json:"spotVisible"`
	} `json:"snapshot"`
	FireTime *ddb.DateOrTime `json:"fire,omitempty"`
}

func Test_API_Dbevent_Unmarshal(test *testing.T) {
	assert := assert.New(test)

	sourceJson := `{"event":{"B":"FyUFoPjuT4iTSHKgIfJ1lg=="},"fire":{"N":"16191216000"},"id":{"B":"WJ+hqZMXRfuNuwU2GEXvog=="},"partition":{"B":"dU9MP7gMrEiCl+e1BtIt78E="},"rights":{"M":{"spot":{"M":{"base":{"N":"700"}}}}},"snapshot":{"M":{"desc":{"S":"Classical"},"loc":{"M":{"latLng":{"L":[{"N":"55.012302"},{"N":"82.92438"}]},"name":{"L":[{"S":"Памятник императору Александру III"},{"S":"Обская ул., Новосибирск, Russia"}]}}},"spotVisible":{"N":"700"},"start":{"N":"16191216000"}}},"spot":{"B":"XQDi12OcRiW0tRxBslkUzQ=="},"user":{"B":"Bfyj8A+WRBCrBlIs10c2/w=="},"ver":{"N":"1"}}`

	source := map[string]events.DynamoDBAttributeValue{}
	assert.NoError(
		json.Unmarshal(
			[]byte(sourceJson),
			&source))

	var record dbeventUnmarshalTestRecord
	assert.NoError(
		apidbevent.UnmarshalEventsDynamoDBAttributeValues(
			source,
			&record))

	// It's not required to compare all fields, just several to see that
	// it has wrote somthing into result.
	assert.NotNil(record.Snapshot.Location)
	assert.Equal(2, len(record.Snapshot.Location.Name))
	assert.Equal(
		"Памятник императору Александру III",
		record.Snapshot.Location.Name[0])
	assert.Equal(
		"Обская ул., Новосибирск, Russia",
		record.Snapshot.Location.Name[1])
}
