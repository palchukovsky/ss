// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package apigateway

import (
	"fmt"
	"regexp"
	"strconv"

	"github.com/palchukovsky/ss"
)

func CheckClientVersionActuality(source string) (bool, error) {
	if appVersionParseRegexp == nil {
		appVersionParseRegexp = regexp.MustCompile(
			`^([0-9]+)\.([0-9]+)\.([0-9]+)\.([0-9]+)$`)
	}

	match := appVersionParseRegexp.FindStringSubmatch(source)
	if len(match) == 0 {
		return false, fmt.Errorf(`failed to parse client version %q`, source)
	}

	comparator := minAppVersionComparator{
		minVersion: ss.S.Config().App.MinVersion,
		appVersion: match,
	}
	return comparator.Compare(), nil
}

////////////////////////////////////////////////////////////////////////////////

var appVersionParseRegexp *regexp.Regexp

type minAppVersionComparator struct {
	minVersion [4]uint
	appVersion []string
}

func (c minAppVersionComparator) Compare() bool {
	var index int
	compare := func() int {
		if index >= len(c.minVersion) {
			return 1
		}
		if v := c.parsePart(index); v < c.minVersion[index] {
			return -1
		} else if v > c.minVersion[index] {
			return 1
		}
		index++
		return 0
	}

	for {
		if r := compare(); r != 0 {
			return r > 0
		}
	}
}

func (c minAppVersionComparator) parsePart(index int) uint {
	ver, err := strconv.ParseUint(c.appVersion[1+index], 10, 32)
	if err != nil {
		ss.S.Log().Panic(
			ss.NewLogMsg(
				`failed to parse version part for %q`,
				c.appVersion[1+index]))
	}
	return uint(ver)
}
