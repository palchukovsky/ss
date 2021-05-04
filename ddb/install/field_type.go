// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ddbinstall

import (
	"reflect"
	"strings"

	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/palchukovsky/ss"
	"github.com/palchukovsky/ss/ddb"
)

func getTypeFields(
	record ddb.IndexRecord,
	source reflect.Type,
	names map[string]struct{},
) {
	if source.Kind() == reflect.Ptr {
		source = source.Elem()
	}
	if source.Kind() != reflect.Struct {
		return
	}

	for i := 0; i < source.NumField(); i++ {
		field := source.Field(i)
		tag, isContinued := getFieldName(field)
		if tag == "" {
			if isContinued {
				getTypeFields(record, field.Type, names)
			}
			continue
		}
		if tag == record.GetKeyPartitionField() {
			continue
		}
		if tag == record.GetKeySortField() {
			continue
		}
		if tag == record.GetIndexPartitionField() {
			continue
		}
		if tag == record.GetIndexSortField() {
			continue
		}
		names[tag] = struct{}{}
	}
}

func getFieldName(field reflect.StructField) (string, bool) {
	tag := field.Tag.Get("json")
	if tag == "" {
		return "", true
	}
	if tag == "-" {
		return "", false
	}
	if commaIdx := strings.Index(tag, ","); commaIdx > 0 {
		tag = tag[:commaIdx]
	}
	return tag, true
}

func getFiledType(record ddb.DataRecord, fieldName string) string {
	result := getTypeType(reflect.ValueOf(record.GetData()).Type(), fieldName)
	if result == "" {
		ss.S.Log().Panic(
			ss.NewLogMsg(
				"failed to find filed type for field %q in table %q",
				record.GetTable(),
				fieldName))
	}
	return result
}

func getTypeType(source reflect.Type, fieldName string) string {
	if source.Kind() == reflect.Ptr {
		source = source.Elem()
	}
	if source.Kind() != reflect.Struct {
		return ""
	}
	for i := 0; i < source.NumField(); i++ {
		field := source.Field(i)
		tag, isContinued := getFieldName(field)
		if tag == "" {
			if isContinued {
				result := getTypeType(field.Type, fieldName)
				if result != "" {
					return result
				}
			}
			continue
		}
		if tag == fieldName {
			return getTypeByType(field.Type)
		}
	}
	return ""
}

func getTypeByType(source reflect.Type) string {
	if source.Kind() == reflect.Ptr {
		source = source.Elem()
	}

	switch source.Kind() {
	case reflect.Bool, reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32,
		reflect.Int64, reflect.Uint, reflect.Uint8, reflect.Uint16,
		reflect.Uint32, reflect.Uint64:
		{
			return dynamodb.ScalarAttributeTypeN
		}
	case reflect.String:
		return dynamodb.ScalarAttributeTypeS
	case reflect.Array:
		return dynamodb.ScalarAttributeTypeB
	}

	switch source.PkgPath() {
	case "github.com/palchukovsky/ss":
		switch source.Name() {
		case "EntityID":
			return dynamodb.ScalarAttributeTypeB
		}
	case "github.com/palchukovsky/ss/ddb":
		switch source.Name() {
		case "Time", "DateOrTime":
			return dynamodb.ScalarAttributeTypeN
		}
	}

	if result := getTypeBySubtype(source); result != "" {
		return result
	}

	ss.S.Log().Panic(
		ss.NewLogMsg(
			"failed to find DB filed type for type %q/%q",
			source.PkgPath(),
			source.Name()))
	return ""
}

func getTypeBySubtype(source reflect.Type) string {
	if source.Kind() == reflect.Ptr {
		source = source.Elem()
	}
	for i := 0; i < source.NumField(); i++ {
		if result := getTypeByType(source.Field(i).Type); result != "" {
			return result
		}
	}
	return ""
}
