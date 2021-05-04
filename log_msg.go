// Copyright 2021, the SS project owners. All rights reserved.
// Please see the OWNERS and LICENSE files for details.

package ss

import (
	"bufio"
	"encoding/json"
	"fmt"
	"reflect"
	"runtime"
	"strings"
	"time"
)

const (
	logMsgNodeTime         = "time"
	logMsgNodeMessage      = "message"
	logMsgNodeLevel        = "level"
	logMsgNodeErrorList    = "errs"
	logMsgNodeUser         = "user"
	logMsgNodeConnection   = "connection"
	logMsgNodeRequest      = "request"
	logMsgNodeRequestList  = "requests"
	logMsgNodeResponseList = "responses"
	logMsgNodeStack        = "stack"
	logMsgNodeMarshalError = "marshalErr"
)

////////////////////////////////////////////////////////////////////////////////

type LogMsg struct {
	message    string
	attributes []LogMsgAttr
	level      logLevel
	errs       []logMsgAttrErr
	time       time.Time
}

func NewLogMsg(format string, args ...interface{}) *LogMsg {
	return &LogMsg{
		message: fmt.Sprintf(format, args...),
		time:    Now(),
	}
}

func (m *LogMsg) AddPrefix(prefix LogPrefix) *LogMsg {
	// Source attributes has to be last to overwired conflicts.
	m.attributes = append(prefix, m.attributes...)
	return m
}

func (m *LogMsg) Add(source LogMsgAttr) *LogMsg {
	m.attributes = append(m.attributes, source)
	return m
}

func (m *LogMsg) AddRequest(source interface{}) *LogMsg {
	m.attributes = append(
		m.attributes,
		NewLogMsgAttrDump(logMsgNodeRequestList, source))
	return m
}
func (m *LogMsg) AddResponse(source interface{}) *LogMsg {
	m.attributes = append(
		m.attributes,
		NewLogMsgAttrDump(logMsgNodeResponseList, source))
	return m
}

func (m *LogMsg) AddErr(source error) *LogMsg {
	m.errs = append(m.errs, newLogMsgAttrErr(source))
	return m
}

func (m *LogMsg) AddUser(source UserID) *LogMsg {
	return m.AddVal(logMsgNodeUser, source)
}

func (m *LogMsg) AddConnectionID(source ConnectionID) *LogMsg {
	return m.AddVal(logMsgNodeConnection, source)
}

func (m *LogMsg) AddRequestID(source string) *LogMsg {
	return m.AddVal(logMsgNodeConnection, source)
}

func (m *LogMsg) AddVal(name string, value interface{}) *LogMsg {
	m.attributes = append(m.attributes, NewLogMsgAttrVal(name, value))
	return m
}

func (m *LogMsg) AddCurrentStack() *LogMsg {
	m.attributes = append(m.attributes, newLogMsgAttrCurrentStack())
	return m
}

func (m *LogMsg) SetLevel(source logLevel) { m.level = source }
func (m LogMsg) GetLevel() logLevel        { return m.level }
func (m LogMsg) GetMessage() string        { return m.message }
func (m LogMsg) GetErrs() []logMsgAttrErr  { return m.errs }

func (m LogMsg) MarshalMap() map[string]interface{} {
	result := m.MarshalAttributesMap()
	result[logMsgNodeMessage] = m.GetMessage()
	result[logMsgNodeLevel] = m.level
	return result
}

func (m LogMsg) ConvertToJSON() []byte {
	result, err := json.Marshal(m.MarshalMap())
	if err != nil {
		result = []byte(
			fmt.Sprintf(
				`{%q:%q}`,
				logMsgNodeMarshalError,
				fmt.Sprintf("%v", err)))
	}
	return result
}

func (m LogMsg) MarshalAttributesMap() map[string]interface{} {
	result := map[string]interface{}{
		logMsgNodeTime: m.time,
	}

	for _, a := range m.attributes {
		a.Marshal(result)
	}

	if m.errs != nil {
		errs := make([]interface{}, len(m.errs))
		for i, e := range m.errs {
			errs[i] = e.Marshal()
		}
		result[logMsgNodeErrorList] = errs
	}

	return result
}

func (m LogMsg) ConvertAttributesToJSON() []byte {
	result, err := json.Marshal(m.MarshalAttributesMap())
	if err != nil {
		result = []byte(
			fmt.Sprintf(
				`{%q:%q}`,
				logMsgNodeMarshalError,
				fmt.Sprintf("%v", err)))
	}
	return result
}

////////////////////////////////////////////////////////////////////////////////

type LogMsgAttr interface {
	Marshal(destination map[string]interface{})
}

////////////////////////////////////////////////////////////////////////////////

type LogMsgAttrVal struct {
	node  string
	value interface{}
}

func NewLogMsgAttrVal(node string, value interface{}) LogMsgAttrVal {
	return LogMsgAttrVal{
		node:  node,
		value: value,
	}
}

func (a LogMsgAttrVal) Marshal(destination map[string]interface{}) {
	destination[a.node] = a.value
}

////////////////////////////////////////////////////////////////////////////////

type LogMsgAttrDump struct {
	node  string
	value interface{}
}

func NewLogMsgAttrDump(node string, obj interface{}) LogMsgAttrDump {
	return LogMsgAttrDump{
		node:  node,
		value: obj,
	}
}

type logMsgAttrDumpValue struct {
	Type  string      `json:"type"`
	Value interface{} `json:"value"`
}

func (a LogMsgAttrDump) Marshal(destination map[string]interface{}) {
	value := logMsgAttrDumpValue{
		Type:  newLogMsgValueTypeName(a.value),
		Value: a.value,
	}

	if node, has := destination[a.node]; has {
		destination[a.node] = append(node.([]logMsgAttrDumpValue), value)
		return
	}
	destination[a.node] = []logMsgAttrDumpValue{value}
}

////////////////////////////////////////////////////////////////////////////////

type logMsgAttrErr struct{ value error }

func newLogMsgAttrErr(source error) logMsgAttrErr {
	return logMsgAttrErr{value: source}
}

func (a logMsgAttrErr) Get() error { return a.value }

func (a logMsgAttrErr) Marshal() interface{} {
	return struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	}{
		Type:  newLogMsgValueTypeName(a.value),
		Value: fmt.Sprintf("%v", a.value),
	}
}

////////////////////////////////////////////////////////////////////////////////

type logMsgAttrStack []interface{}

func newLogMsgAttrCurrentStack() logMsgAttrStack {

	buffer := make([]byte, 4096)
	size := runtime.Stack(buffer, false)

	scanner := bufio.NewScanner(strings.NewReader(string(buffer[:size])))

	stack := []interface{}{}
	isFirst := true
	for scanner.Scan() {
		if isFirst {
			// Skipping "goroutine 1 [running]:"...
			isFirst = false
			continue
		}

		level := struct {
			Func string `json:"func"`
			Path string `json:"path"`
		}{
			Func: scanner.Text(),
		}
		if scanner.Scan() {
			level.Path = strings.Trim(scanner.Text(), "\t\n\r ")
		}

		stack = append(stack, level)
	}

	return logMsgAttrStack(stack)
}

func (a logMsgAttrStack) Marshal(destination map[string]interface{}) {
	destination[logMsgNodeStack] = a
}

////////////////////////////////////////////////////////////////////////////////

type LogPrefix []LogMsgAttr

func NewLogPrefix() LogPrefix { return make([]LogMsgAttr, 0, 1) }

func (lp LogPrefix) Add(a LogMsgAttr) LogPrefix {
	lp = append(lp, a)
	return lp
}

func (lp LogPrefix) AddUser(source UserID) LogPrefix {
	return lp.AddVal(logMsgNodeUser, source)
}

func (lp LogPrefix) AddConnectionID(source ConnectionID) LogPrefix {
	return lp.AddVal(logMsgNodeConnection, source)
}

func (lp LogPrefix) AddRequestID(source string) LogPrefix {
	return lp.AddVal(logMsgNodeRequest, source)
}

func (lp LogPrefix) AddVal(name string, value interface{}) LogPrefix {
	lp = append(lp, NewLogMsgAttrVal(name, value))
	return lp
}

////////////////////////////////////////////////////////////////////////////////

func newLogMsgValueTypeName(source interface{}) string {
	result := ""
	t := reflect.TypeOf(source)
	isPtr := t.Kind() == reflect.Ptr
	if isPtr {
		t = t.Elem()
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

////////////////////////////////////////////////////////////////////////////////
