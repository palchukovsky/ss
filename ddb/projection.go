// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ddb

import (
	"reflect"
	"strings"

	"github.com/aws/aws-sdk-go/aws"
)

func getRecordProjection(record RecordBuffer, aliases *map[string]*string,
) *string {
	var result string
	getTypeProjection(reflect.ValueOf(record).Type(), &result, aliases)
	return aws.String(result)
}

func getTypeProjection(
	source reflect.Type, projection *string, aliases *map[string]*string,
) {
	if source.Kind() == reflect.Ptr {
		source = source.Elem()
	}
	if source.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < source.NumField(); i++ {
		field := source.Field(i)
		tag := field.Tag.Get("json")
		if tag == "" {
			getTypeProjection(field.Type, projection, aliases)
			continue
		}
		if tag == "-" {
			continue
		}

		if commaIdx := strings.Index(tag, ","); commaIdx > 0 {
			tag = tag[:commaIdx]
		}

		if isReservedWord(tag) {
			alias := "#" + tag
			if *aliases == nil {
				*aliases = map[string]*string{alias: aws.String(tag)}
			} else {
				(*aliases)[alias] = aws.String(tag)
			}
			tag = alias
		}

		if *projection != "" {
			*projection += ","
		}
		*projection += tag
	}
}
