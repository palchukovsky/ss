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
	logMsgNodeTime                  = "time"
	logMsgNodeMessage               = "message"
	logMsgNodeLevel                 = "level"
	logMsgNodeErrorList             = "err"
	logMsgNodeUser                  = "user"
	logMsgNodeConnection            = "connection"
	logMsgNodeRequest               = "request"
	logMsgNodeDumpList              = "dump"
	logMsgNodeDumpGroupRequestList  = "request"
	logMsgNodeDumpGroupResponseList = "request"
	logMsgNodeStack                 = "stack"
	logMsgNodeMarshalError          = "marshalErr"
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
		NewLogMsgAttrDumpGroup(logMsgNodeDumpGroupRequestList, source))
	return m
}
func (m *LogMsg) AddResponse(source interface{}) *LogMsg {
	m.attributes = append(
		m.attributes,
		NewLogMsgAttrDumpGroup(logMsgNodeDumpGroupResponseList, source))
	return m
}

func (m *LogMsg) AddErr(source error) *LogMsg {
	m.errs = append(m.errs, newLogMsgAttrError(source))
	return m
}

func (m *LogMsg) AddPanic(source interface{}) *LogMsg {
	m.errs = append(m.errs, newlogMsgAttrPanic(source))
	return m
}

func (m *LogMsg) AddRequestID(source string) *LogMsg {
	return m.AddVal(logMsgNodeConnection, source)
}

func (m *LogMsg) AddDump(source interface{}) *LogMsg {
	m.attributes = append(m.attributes, newLogMsgAttrDump(source))
	return m
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
	value := m.MarshalMap()
	result, err := json.Marshal(value)
	if err != nil {
		result = []byte(
			fmt.Sprintf(
				`{%q:{"err":%q,"value":%q}}`,
				logMsgNodeMarshalError,
				fmt.Sprintf("%v", err),
				fmt.Sprintf("%v", value)))
	}
	return result
}

func (m LogMsg) MarshalAttributesMap() map[string]interface{} {
	result := map[string]interface{}{
		logMsgNodeTime: m.time,
	}

	for _, a := range m.attributes {
		a.MarshalLogMsg(result)
	}

	if m.errs != nil {
		errs := make([]interface{}, len(m.errs))
		for i, e := range m.errs {
			errs[i] = e.MarshalLogMsg()
		}
		result[logMsgNodeErrorList] = errs
	}

	return result
}

func (m LogMsg) ConvertAttributesToJSON() []byte {
	value := m.MarshalAttributesMap()
	result, err := json.Marshal(value)
	if err != nil {
		result = []byte(
			fmt.Sprintf(
				`{%q:{"err":%q,"value":%q}}`,
				logMsgNodeMarshalError,
				fmt.Sprintf("%v", err),
				fmt.Sprintf("%#v", value)))
	}
	return result
}

////////////////////////////////////////////////////////////////////////////////

type LogMsgAttr interface {
	MarshalLogMsg(destination map[string]interface{})
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

func (a LogMsgAttrVal) MarshalLogMsg(destination map[string]interface{}) {
	destination[a.node] = a.value
}

////////////////////////////////////////////////////////////////////////////////

type logMsgAttrDump struct{ value interface{} }

func newLogMsgAttrDump(value interface{}) logMsgAttrDump {
	return logMsgAttrDump{value: value}
}

func MarshalLogMsgAttrDump(
	value interface{},
	destination map[string]interface{},
) {
	marshalValue := logMsgAttrDumpValue{newLogMsgValueTypeName(value): value}
	if node, has := destination[logMsgNodeDumpList]; has {
		destination[logMsgNodeDumpList] = append(
			node.([]logMsgAttrDumpValue),
			marshalValue)
		return
	}
	destination[logMsgNodeDumpList] = []logMsgAttrDumpValue{marshalValue}
}

func (a logMsgAttrDump) MarshalLogMsg(destination map[string]interface{}) {
	MarshalLogMsgAttrDump(a.value, destination)
}

type logMsgAttrDumpValue map[string]interface{}

type LogMsgAttrDumpGroup struct {
	logMsgAttrDump
	groupNode string
}

func NewLogMsgAttrDumpGroup(
	groupNode string,
	value interface{},
) LogMsgAttrDumpGroup {
	return LogMsgAttrDumpGroup{
		logMsgAttrDump: newLogMsgAttrDump(value),
		groupNode:      groupNode,
	}
}

func (a LogMsgAttrDumpGroup) MarshalLogMsg(destination map[string]interface{}) {
	value := logMsgAttrDumpValue{newLogMsgValueTypeName(a.value): a.value}
	if dumps, has := destination[logMsgNodeDumpList]; has {
		destination := dumps.(map[string]interface{})
		if node, has := destination[a.groupNode]; has {
			destination[a.groupNode] = append(
				node.([]logMsgAttrDumpValue),
				value)
			return
		}
		destination[a.groupNode] = []logMsgAttrDumpValue{value}
		return
	}
	destination[logMsgNodeDumpList] = map[string]interface{}{
		a.groupNode: []logMsgAttrDumpValue{value},
	}
}

////////////////////////////////////////////////////////////////////////////////

type logMsgAttrErr interface {
	Get() error
	MarshalLogMsg() interface{}
}

type logMsgAttrError struct{ value error }

func newLogMsgAttrError(source error) logMsgAttrErr {
	return logMsgAttrError{value: source}
}

func (a logMsgAttrError) Get() error { return a.value }

func (a logMsgAttrError) MarshalLogMsg() interface{} {
	return struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	}{
		Type:  newLogMsgValueTypeName(a.value),
		Value: a.Get().Error(),
	}
}

type logMsgAttrPanic struct{ value logMsgAttrPanicValue }

func newlogMsgAttrPanic(source interface{}) logMsgAttrErr {
	return logMsgAttrPanic{value: logMsgAttrPanicValue{value: source}}
}

func (a logMsgAttrPanic) Get() error { return a.value }

func (a logMsgAttrPanic) MarshalLogMsg() interface{} {
	return struct {
		Type  string `json:"type"`
		Value string `json:"value"`
	}{
		Type:  newLogMsgValueTypeName(a.value),
		Value: a.Get().Error(),
	}
}

type logMsgAttrPanicValue struct{ value interface{} }

func (v logMsgAttrPanicValue) Error() string {
	return fmt.Sprintf("%v", v.value)
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

func (a logMsgAttrStack) MarshalLogMsg(destination map[string]interface{}) {
	destination[logMsgNodeStack] = a
}

////////////////////////////////////////////////////////////////////////////////

type LogPrefix []LogMsgAttr

func NewLogPrefix() LogPrefix { return make([]LogMsgAttr, 0, 1) }

func (lp LogPrefix) Add(a LogMsgAttr) LogPrefix {
	lp = append(lp, a)
	return lp
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
	if source == nil {
		return "nil"
	}
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
