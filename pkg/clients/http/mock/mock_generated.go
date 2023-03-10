// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/antonpriyma/otus-highload/pkg/clients/http (interfaces: Client)

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	http "github.com/antonpriyma/otus-highload/pkg/clients/http"
	reflect "reflect"
)

// MockClient is a mock of Client interface
type MockClient struct {
	ctrl     *gomock.Controller
	recorder *MockClientMockRecorder
}

// MockClientMockRecorder is the mock recorder for MockClient
type MockClientMockRecorder struct {
	mock *MockClient
}

// NewMockClient creates a new mock instance
func NewMockClient(ctrl *gomock.Controller) *MockClient {
	mock := &MockClient{ctrl: ctrl}
	mock.recorder = &MockClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use
func (m *MockClient) EXPECT() *MockClientMockRecorder {
	return m.recorder
}

// PerformRequest mocks base method
func (m *MockClient) PerformRequest(arg0 context.Context, arg1 http.Request, arg2 http.Response) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "PerformRequest", arg0, arg1, arg2)
	ret0, _ := ret[0].(error)
	return ret0
}

// PerformRequest indicates an expected call of PerformRequest
func (mr *MockClientMockRecorder) PerformRequest(arg0, arg1, arg2 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PerformRequest", reflect.TypeOf((*MockClient)(nil).PerformRequest), arg0, arg1, arg2)
}
