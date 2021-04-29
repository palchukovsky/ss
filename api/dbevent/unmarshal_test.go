// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package apidbevent_test

import (
	"encoding/json"
	"testing"

	"github.com/aws/aws-lambda-go/events"
	apidbevent "github.com/palchukovsky/ss/api/dbevent"
	"github.com/palchukovsky/ss/ddb"
	"github.com/stretchr/testify/assert"
)

type dbeventUnmarshalTestRecord struct {
	/*ID        ss.EntityID  `json:"id"`
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
	} `json:"snapshot"`*/
	FireTime *ddb.DateOrTime `json:"fire,omitempty"`
}

func Test_API_Dbevent_Unmarshal(test *testing.T) {
	assert := assert.New(test)

	sourceJson := `{"event":{"B":"qGq9X5sxRiel/FK5Y0kKGA=="},"fire":{"N":"16197525000"},"id":{"B":"3twB+YpHTimJj3a/+7G1ew=="},"partition":{"B":"dRATLRiFHEBOk+KPmWTJk+OPYMLb5XVMW7EeUj8Fx3Q8"},"rights":{"M":{"spot":{"M":{"base":{"N":"100"}}}}},"snapshot":{"M":{"desc":{"S":"Edited event description 123@"},"loc":{"M":{"latLng":{"L":[{"N":"48.218597"},{"N":"16.363022"}]},"name":{"L":[{"S":"Berggasse 19"},{"S":"Vienna, Austria"}]}}},"spotVisible":{"N":"700"},"start":{"N":"16197525000"}}},"spot":{"B":"IutVALucTQymUncp7AegSA=="},"user":{"B":"EBMtGIUcQE6T4o+ZZMmT4w=="},"ver":{"N":"2"}}`

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

	assert.Empty(record.FireTime.Value.Get().String())

}
