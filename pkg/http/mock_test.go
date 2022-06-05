// Code generated by MockGen. DO NOT EDIT.
// Source: ./stale-while-revalidate.go

// Package http is a generated GoMock package.
package http

import (
	context "context"
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	model "neurocode.io/cache-offloader/pkg/model"
)

// MockCacher is a mock of Cacher interface.
type MockCacher struct {
	ctrl     *gomock.Controller
	recorder *MockCacherMockRecorder
}

// MockCacherMockRecorder is the mock recorder for MockCacher.
type MockCacherMockRecorder struct {
	mock *MockCacher
}

// NewMockCacher creates a new mock instance.
func NewMockCacher(ctrl *gomock.Controller) *MockCacher {
	mock := &MockCacher{ctrl: ctrl}
	mock.recorder = &MockCacherMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockCacher) EXPECT() *MockCacherMockRecorder {
	return m.recorder
}

// LookUp mocks base method.
func (m *MockCacher) LookUp(arg0 context.Context, arg1 string) (*model.Response, error) {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "LookUp", arg0, arg1)
	ret0, _ := ret[0].(*model.Response)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// LookUp indicates an expected call of LookUp.
func (mr *MockCacherMockRecorder) LookUp(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "LookUp", reflect.TypeOf((*MockCacher)(nil).LookUp), arg0, arg1)
}

// Store mocks base method.
func (m *MockCacher) Store(arg0 context.Context, arg1 string, arg2 *model.Response) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "Store", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// Store indicates an expected call of Store.
func (mr *MockCacherMockRecorder) Store(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Store", reflect.TypeOf((*MockCacher)(nil).Store), arg0, arg1, arg2)
}

// MockMetricsCollector is a mock of MetricsCollector interface.
type MockMetricsCollector struct {
	ctrl     *gomock.Controller
	recorder *MockMetricsCollectorMockRecorder
}

// MockMetricsCollectorMockRecorder is the mock recorder for MockMetricsCollector.
type MockMetricsCollectorMockRecorder struct {
	mock *MockMetricsCollector
}

// NewMockMetricsCollector creates a new mock instance.
func NewMockMetricsCollector(ctrl *gomock.Controller) *MockMetricsCollector {
	mock := &MockMetricsCollector{ctrl: ctrl}
	mock.recorder = &MockMetricsCollectorMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockMetricsCollector) EXPECT() *MockMetricsCollectorMockRecorder {
	return m.recorder
}

// CacheHit mocks base method.
func (m *MockMetricsCollector) CacheHit(method string, statusCode int) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "CacheHit", method, statusCode)
}

// CacheHit indicates an expected call of CacheHit.
func (mr *MockMetricsCollectorMockRecorder) CacheHit(method, statusCode interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CacheHit", reflect.TypeOf((*MockMetricsCollector)(nil).CacheHit), method, statusCode)
}

// CacheMiss mocks base method.
func (m *MockMetricsCollector) CacheMiss(method string, statusCode int) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "CacheMiss", method, statusCode)
}

// CacheMiss indicates an expected call of CacheMiss.
func (mr *MockMetricsCollectorMockRecorder) CacheMiss(method, statusCode interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "CacheMiss", reflect.TypeOf((*MockMetricsCollector)(nil).CacheMiss), method, statusCode)
}