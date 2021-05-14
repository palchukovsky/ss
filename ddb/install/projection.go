// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ddbinstall

import (
	"reflect"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/palchukovsky/ss"
	"github.com/palchukovsky/ss/ddb"
)

func getIndexProjection(records ...ddb.IndexRecord) *dynamodb.Projection {
	var table string
	var index string

	names := map[string]struct{}{}
	for _, record := range records {

		if table != record.GetTable() {
			if table != "" {
				ss.S.Log().Panic(
					ss.NewLogMsg(
						"wrong table to get index field list, was %q but now %q",
						table,
						record.GetTable()))
			}
			table = record.GetTable()
		}
		if index != record.GetIndex() {
			if index != "" {
				ss.S.Log().Panic(
					ss.NewLogMsg(
						"wrong table %q index to get index field list, was %q but now %q",
						table,
						index,
						record.GetIndex()))
			}
			index = record.GetIndex()
		}

		getTypeFields(record, reflect.ValueOf(record).Type(), names)

		for _, name := range record.GetProjection() {
			names[name] = struct{}{}
		}
	}

	if len(names) == 0 {
		return &dynamodb.Projection{
			ProjectionType: aws.String(dynamodb.ProjectionTypeKeysOnly),
		}
	}
	result := dynamodb.Projection{
		ProjectionType:   aws.String(dynamodb.ProjectionTypeInclude),
		NonKeyAttributes: make([]*string, 0, len(names)),
	}
	for name := range names {
		result.NonKeyAttributes = append(result.NonKeyAttributes, aws.String(name))
	}
	return &result
}
