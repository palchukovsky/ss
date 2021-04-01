// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ddb

import (
	"regexp"

	"github.com/aws/aws-sdk-go/aws"
)

func aliasReservedInString(source string, aliases *map[string]*string,
) *string {
	matches := getAliasReserverdInStringRegexp().FindAllStringIndex(source, -1)
	if len(matches) == 0 {
		return aws.String(source)
	}

	ranges := make([][]int, 0, len(matches))
	check := func(beg, end int) {
		if !isReservedWord(source[beg:end]) {
			return
		}
		ranges = append(ranges, []int{beg, end})
	}

	beg := 0
	end := 0
	for _, match := range matches {
		end = match[0]
		if match[1] != 0 {
			check(beg, end)
		}
		beg = match[1]
	}
	if end != len(source) {
		check(beg, len(source))
	}

	for i := len(ranges) - 1; i >= 0; i-- {
		r := ranges[i]
		w := source[r[0]:r[1]]
		if *aliases == nil {
			*aliases = make(map[string]*string)
		}
		(*aliases)["#"+w] = aws.String(w)
		source = source[:r[0]] + "#" + source[r[0]:]
	}
	return aws.String(source)
}

func aliasReservedWord(source string, aliases *map[string]*string,
) string {
	if !isReservedWord(source) {
		return source
	}
	if *aliases == nil {
		*aliases = make(map[string]*string)
	}
	result := "#" + source
	(*aliases)[result] = aws.String(source)
	return result
}

func isReservedWord(source string) bool {
	switch source {
	case "user", "snapshot", "next", "name", "token",
		"type", "desc", "start", "partition":
		return true
	default:
		return false
	}

}

var aliasReserverdInStringRegexp *regexp.Regexp

func getAliasReserverdInStringRegexp() *regexp.Regexp {
	if aliasReserverdInStringRegexp == nil {
		aliasReserverdInStringRegexp = regexp.MustCompile(`[^\w\d]`)
	}
	return aliasReserverdInStringRegexp
}
