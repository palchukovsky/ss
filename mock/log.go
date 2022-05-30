// Code generated by MockGen. DO NOT EDIT.
// Source: log.go

// Package mock_ss is a generated GoMock package.
package mock_ss

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	ss "github.com/palchukovsky/ss"
)

// MockLogSource is a mock of LogSource interface.
type MockLogSource struct {
	ctrl     *gomock.Controller
	recorder *MockLogSourceMockRecorder
}

// MockLogSourceMockRecorder is the mock recorder for MockLogSource.
type MockLogSourceMockRecorder struct {
	mock *MockLogSource
}

// NewMockLogSource creates a new mock instance.
func NewMockLogSource(ctrl *gomock.Controller) *MockLogSource {
	mock := &MockLogSource{ctrl: ctrl}
	mock.recorder = &MockLogSourceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockLogSource) EXPECT() *MockLogSourceMockRecorder {
	return m.recorder
}

// Log mocks base method.
func (m *MockLogSource) Log() ss.LogStream {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Log")
	ret0, _ := ret[0].(ss.LogStream)
	return ret0
}

// Log indicates an expected call of Log.
func (mr *MockLogSourceMockRecorder) Log() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Log", reflect.TypeOf((*MockLogSource)(nil).Log))
}

// MockLogStream is a mock of LogStream interface.
type MockLogStream struct {
	ctrl     *gomock.Controller
	recorder *MockLogStreamMockRecorder
}

// MockLogStreamMockRecorder is the mock recorder for MockLogStream.
type MockLogStreamMockRecorder struct {
	mock *MockLogStream
}

// NewMockLogStream creates a new mock instance.
func NewMockLogStream(ctrl *gomock.Controller) *MockLogStream {
	mock := &MockLogStream{ctrl: ctrl}
	mock.recorder = &MockLogStreamMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockLogStream) EXPECT() *MockLogStreamMockRecorder {
	return m.recorder
}

// CheckPanic mocks base method.
func (m *MockLogStream) CheckPanic(panicValue interface{}, errorMessage string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "CheckPanic", panicValue, errorMessage)
}

// CheckPanic indicates an expected call of CheckPanic.
func (mr *MockLogStreamMockRecorder) CheckPanic(panicValue, errorMessage interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckPanic", reflect.TypeOf((*MockLogStream)(nil).CheckPanic), panicValue, errorMessage)
}

// Debug mocks base method.
func (m *MockLogStream) Debug(arg0 *ss.LogMsg) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Debug", arg0)
}

// Debug indicates an expected call of Debug.
func (mr *MockLogStreamMockRecorder) Debug(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Debug", reflect.TypeOf((*MockLogStream)(nil).Debug), arg0)
}

// Error mocks base method.
func (m *MockLogStream) Error(arg0 *ss.LogMsg) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Error", arg0)
}

// Error indicates an expected call of Error.
func (mr *MockLogStreamMockRecorder) Error(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Error", reflect.TypeOf((*MockLogStream)(nil).Error), arg0)
}

// Info mocks base method.
func (m *MockLogStream) Info(arg0 *ss.LogMsg) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Info", arg0)
}

// Info indicates an expected call of Info.
func (mr *MockLogStreamMockRecorder) Info(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Info", reflect.TypeOf((*MockLogStream)(nil).Info), arg0)
}

// Lock mocks base method.
func (m *MockLogStream) Lock() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Lock")
}

// Lock indicates an expected call of Lock.
func (mr *MockLogStreamMockRecorder) Lock() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Lock", reflect.TypeOf((*MockLogStream)(nil).Lock))
}

// Panic mocks base method.
func (m *MockLogStream) Panic(arg0 *ss.LogMsg) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Panic", arg0)
}

// Panic indicates an expected call of Panic.
func (mr *MockLogStreamMockRecorder) Panic(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Panic", reflect.TypeOf((*MockLogStream)(nil).Panic), arg0)
}

// Unlock mocks base method.
func (m *MockLogStream) Unlock() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Unlock")
}

// Unlock indicates an expected call of Unlock.
func (mr *MockLogStreamMockRecorder) Unlock() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Unlock", reflect.TypeOf((*MockLogStream)(nil).Unlock))
}

// Warn mocks base method.
func (m *MockLogStream) Warn(arg0 *ss.LogMsg) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Warn", arg0)
}

// Warn indicates an expected call of Warn.
func (mr *MockLogStreamMockRecorder) Warn(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Warn", reflect.TypeOf((*MockLogStream)(nil).Warn), arg0)
}

// checkPanic mocks base method.
func (m *MockLogStream) checkPanic(panicValue interface{}, getPanicDetails func() *ss.LogMsg) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "checkPanic", panicValue, getPanicDetails)
}

// checkPanic indicates an expected call of checkPanic.
func (mr *MockLogStreamMockRecorder) checkPanic(panicValue, getPanicDetails interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "checkPanic", reflect.TypeOf((*MockLogStream)(nil).checkPanic), panicValue, getPanicDetails)
}

// MockLog is a mock of Log interface.
type MockLog struct {
	ctrl     *gomock.Controller
	recorder *MockLogMockRecorder
}

// MockLogMockRecorder is the mock recorder for MockLog.
type MockLogMockRecorder struct {
	mock *MockLog
}

// NewMockLog creates a new mock instance.
func NewMockLog(ctrl *gomock.Controller) *MockLog {
	mock := &MockLog{ctrl: ctrl}
	mock.recorder = &MockLogMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockLog) EXPECT() *MockLogMockRecorder {
	return m.recorder
}

// CheckExit mocks base method.
func (m *MockLog) CheckExit(panicValue interface{}) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "CheckExit", panicValue)
}

// CheckExit indicates an expected call of CheckExit.
func (mr *MockLogMockRecorder) CheckExit(panicValue interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckExit", reflect.TypeOf((*MockLog)(nil).CheckExit), panicValue)
}

// CheckPanic mocks base method.
func (m *MockLog) CheckPanic(panicValue interface{}, errorMessage string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "CheckPanic", panicValue, errorMessage)
}

// CheckPanic indicates an expected call of CheckPanic.
func (mr *MockLogMockRecorder) CheckPanic(panicValue, errorMessage interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckPanic", reflect.TypeOf((*MockLog)(nil).CheckPanic), panicValue, errorMessage)
}

// Debug mocks base method.
func (m *MockLog) Debug(arg0 *ss.LogMsg) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Debug", arg0)
}

// Debug indicates an expected call of Debug.
func (mr *MockLogMockRecorder) Debug(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Debug", reflect.TypeOf((*MockLog)(nil).Debug), arg0)
}

// Error mocks base method.
func (m *MockLog) Error(arg0 *ss.LogMsg) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Error", arg0)
}

// Error indicates an expected call of Error.
func (mr *MockLogMockRecorder) Error(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Error", reflect.TypeOf((*MockLog)(nil).Error), arg0)
}

// Info mocks base method.
func (m *MockLog) Info(arg0 *ss.LogMsg) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Info", arg0)
}

// Info indicates an expected call of Info.
func (mr *MockLogMockRecorder) Info(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Info", reflect.TypeOf((*MockLog)(nil).Info), arg0)
}

// Lock mocks base method.
func (m *MockLog) Lock() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Lock")
}

// Lock indicates an expected call of Lock.
func (mr *MockLogMockRecorder) Lock() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Lock", reflect.TypeOf((*MockLog)(nil).Lock))
}

// NewSession mocks base method.
func (m *MockLog) NewSession(newPrefix func() ss.LogPrefix) ss.LogSession {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NewSession", newPrefix)
	ret0, _ := ret[0].(ss.LogSession)
	return ret0
}

// NewSession indicates an expected call of NewSession.
func (mr *MockLogMockRecorder) NewSession(newPrefix interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewSession", reflect.TypeOf((*MockLog)(nil).NewSession), newPrefix)
}

// Panic mocks base method.
func (m *MockLog) Panic(arg0 *ss.LogMsg) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Panic", arg0)
}

// Panic indicates an expected call of Panic.
func (mr *MockLogMockRecorder) Panic(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Panic", reflect.TypeOf((*MockLog)(nil).Panic), arg0)
}

// Started mocks base method.
func (m *MockLog) Started() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Started")
}

// Started indicates an expected call of Started.
func (mr *MockLogMockRecorder) Started() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Started", reflect.TypeOf((*MockLog)(nil).Started))
}

// Unlock mocks base method.
func (m *MockLog) Unlock() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Unlock")
}

// Unlock indicates an expected call of Unlock.
func (mr *MockLogMockRecorder) Unlock() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Unlock", reflect.TypeOf((*MockLog)(nil).Unlock))
}

// Warn mocks base method.
func (m *MockLog) Warn(arg0 *ss.LogMsg) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Warn", arg0)
}

// Warn indicates an expected call of Warn.
func (mr *MockLogMockRecorder) Warn(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Warn", reflect.TypeOf((*MockLog)(nil).Warn), arg0)
}

// checkPanic mocks base method.
func (m *MockLog) checkPanic(panicValue interface{}, getPanicDetails func() *ss.LogMsg) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "checkPanic", panicValue, getPanicDetails)
}

// checkPanic indicates an expected call of checkPanic.
func (mr *MockLogMockRecorder) checkPanic(panicValue, getPanicDetails interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "checkPanic", reflect.TypeOf((*MockLog)(nil).checkPanic), panicValue, getPanicDetails)
}

// MockLogSession is a mock of LogSession interface.
type MockLogSession struct {
	ctrl     *gomock.Controller
	recorder *MockLogSessionMockRecorder
}

// MockLogSessionMockRecorder is the mock recorder for MockLogSession.
type MockLogSessionMockRecorder struct {
	mock *MockLogSession
}

// NewMockLogSession creates a new mock instance.
func NewMockLogSession(ctrl *gomock.Controller) *MockLogSession {
	mock := &MockLogSession{ctrl: ctrl}
	mock.recorder = &MockLogSessionMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockLogSession) EXPECT() *MockLogSessionMockRecorder {
	return m.recorder
}

// CheckPanic mocks base method.
func (m *MockLogSession) CheckPanic(panicValue interface{}, errorMessage string) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "CheckPanic", panicValue, errorMessage)
}

// CheckPanic indicates an expected call of CheckPanic.
func (mr *MockLogSessionMockRecorder) CheckPanic(panicValue, errorMessage interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CheckPanic", reflect.TypeOf((*MockLogSession)(nil).CheckPanic), panicValue, errorMessage)
}

// Debug mocks base method.
func (m *MockLogSession) Debug(arg0 *ss.LogMsg) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Debug", arg0)
}

// Debug indicates an expected call of Debug.
func (mr *MockLogSessionMockRecorder) Debug(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Debug", reflect.TypeOf((*MockLogSession)(nil).Debug), arg0)
}

// Error mocks base method.
func (m *MockLogSession) Error(arg0 *ss.LogMsg) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Error", arg0)
}

// Error indicates an expected call of Error.
func (mr *MockLogSessionMockRecorder) Error(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Error", reflect.TypeOf((*MockLogSession)(nil).Error), arg0)
}

// Info mocks base method.
func (m *MockLogSession) Info(arg0 *ss.LogMsg) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Info", arg0)
}

// Info indicates an expected call of Info.
func (mr *MockLogSessionMockRecorder) Info(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Info", reflect.TypeOf((*MockLogSession)(nil).Info), arg0)
}

// Lock mocks base method.
func (m *MockLogSession) Lock() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Lock")
}

// Lock indicates an expected call of Lock.
func (mr *MockLogSessionMockRecorder) Lock() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Lock", reflect.TypeOf((*MockLogSession)(nil).Lock))
}

// NewSession mocks base method.
func (m *MockLogSession) NewSession(newPrefix func() ss.LogPrefix) ss.LogSession {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "NewSession", newPrefix)
	ret0, _ := ret[0].(ss.LogSession)
	return ret0
}

// NewSession indicates an expected call of NewSession.
func (mr *MockLogSessionMockRecorder) NewSession(newPrefix interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "NewSession", reflect.TypeOf((*MockLogSession)(nil).NewSession), newPrefix)
}

// Panic mocks base method.
func (m *MockLogSession) Panic(arg0 *ss.LogMsg) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Panic", arg0)
}

// Panic indicates an expected call of Panic.
func (mr *MockLogSessionMockRecorder) Panic(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Panic", reflect.TypeOf((*MockLogSession)(nil).Panic), arg0)
}

// Unlock mocks base method.
func (m *MockLogSession) Unlock() {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Unlock")
}

// Unlock indicates an expected call of Unlock.
func (mr *MockLogSessionMockRecorder) Unlock() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Unlock", reflect.TypeOf((*MockLogSession)(nil).Unlock))
}

// Warn mocks base method.
func (m *MockLogSession) Warn(arg0 *ss.LogMsg) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "Warn", arg0)
}

// Warn indicates an expected call of Warn.
func (mr *MockLogSessionMockRecorder) Warn(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Warn", reflect.TypeOf((*MockLogSession)(nil).Warn), arg0)
}

// checkPanic mocks base method.
func (m *MockLogSession) checkPanic(panicValue interface{}, getPanicDetails func() *ss.LogMsg) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "checkPanic", panicValue, getPanicDetails)
}

// checkPanic indicates an expected call of checkPanic.
func (mr *MockLogSessionMockRecorder) checkPanic(panicValue, getPanicDetails interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "checkPanic", reflect.TypeOf((*MockLogSession)(nil).checkPanic), panicValue, getPanicDetails)
}

// MocklogDestination is a mock of logDestination interface.
type MocklogDestination struct {
	ctrl     *gomock.Controller
	recorder *MocklogDestinationMockRecorder
}

// MocklogDestinationMockRecorder is the mock recorder for MocklogDestination.
type MocklogDestinationMockRecorder struct {
	mock *MocklogDestination
}

// NewMocklogDestination creates a new mock instance.
func NewMocklogDestination(ctrl *gomock.Controller) *MocklogDestination {
	mock := &MocklogDestination{ctrl: ctrl}
	mock.recorder = &MocklogDestinationMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MocklogDestination) EXPECT() *MocklogDestinationMockRecorder {
	return m.recorder
}

// GetName mocks base method.
func (m *MocklogDestination) GetName() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "GetName")
	ret0, _ := ret[0].(string)
	return ret0
}

// GetName indicates an expected call of GetName.
func (mr *MocklogDestinationMockRecorder) GetName() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "GetName", reflect.TypeOf((*MocklogDestination)(nil).GetName))
}

// Sync mocks base method.
func (m *MocklogDestination) Sync() error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Sync")
	ret0, _ := ret[0].(error)
	return ret0
}

// Sync indicates an expected call of Sync.
func (mr *MocklogDestinationMockRecorder) Sync() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Sync", reflect.TypeOf((*MocklogDestination)(nil).Sync))
}

// WriteDebug mocks base method.
func (m *MocklogDestination) WriteDebug(arg0 *ss.LogMsg) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WriteDebug", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// WriteDebug indicates an expected call of WriteDebug.
func (mr *MocklogDestinationMockRecorder) WriteDebug(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WriteDebug", reflect.TypeOf((*MocklogDestination)(nil).WriteDebug), arg0)
}

// WriteError mocks base method.
func (m *MocklogDestination) WriteError(arg0 *ss.LogMsg) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WriteError", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// WriteError indicates an expected call of WriteError.
func (mr *MocklogDestinationMockRecorder) WriteError(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WriteError", reflect.TypeOf((*MocklogDestination)(nil).WriteError), arg0)
}

// WriteInfo mocks base method.
func (m *MocklogDestination) WriteInfo(arg0 *ss.LogMsg) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WriteInfo", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// WriteInfo indicates an expected call of WriteInfo.
func (mr *MocklogDestinationMockRecorder) WriteInfo(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WriteInfo", reflect.TypeOf((*MocklogDestination)(nil).WriteInfo), arg0)
}

// WritePanic mocks base method.
func (m *MocklogDestination) WritePanic(arg0 *ss.LogMsg) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WritePanic", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// WritePanic indicates an expected call of WritePanic.
func (mr *MocklogDestinationMockRecorder) WritePanic(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WritePanic", reflect.TypeOf((*MocklogDestination)(nil).WritePanic), arg0)
}

// WriteWarn mocks base method.
func (m *MocklogDestination) WriteWarn(arg0 *ss.LogMsg) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "WriteWarn", arg0)
	ret0, _ := ret[0].(error)
	return ret0
}

// WriteWarn indicates an expected call of WriteWarn.
func (mr *MocklogDestinationMockRecorder) WriteWarn(arg0 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "WriteWarn", reflect.TypeOf((*MocklogDestination)(nil).WriteWarn), arg0)
}
