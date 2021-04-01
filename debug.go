// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ss

import (
	"encoding/json"
	"fmt"
	"reflect"
)

// Dump tries to convert object to string for loggin.
func Dump(source interface{}) string {
	result, err := json.Marshal(source)
	if err != nil {
		return fmt.Sprintf(`<<<%s: %v>>>`, GetTypeName(source), source)
	}
	return "<<<" + GetTypeName(source) + ": " + string(result) + ">>>"
}

// GetTypeName returns variable type name.
func GetTypeName(source interface{}) string {
	result := ""
	t := reflect.TypeOf(source)
	isPtr := t.Kind() == reflect.Ptr
	if isPtr {
		t = t.Elem()
		result += "*"
	}
	{
		pkg := t.PkgPath()
		if len(pkg) != 0 {
			result += pkg + "/"
		}
	}
	result += t.Name()
	return result
}
