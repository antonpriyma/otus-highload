// Code generated by MockGen. DO NOT EDIT.
// Source: github.com/antonpriyma/otus-highload/pkg/clients/dwh (interfaces: Client)

// Package mock is a generated GoMock package.
package mock

import (
	context "context"
	gomock "github.com/golang/mock/gomock"
	dwh "github.com/antonpriyma/otus-highload/pkg/clients/dwh"
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

// SendCompactMessage mocks base method
func (m *MockClient) SendCompactMessage(arg0 context.Context, arg1 dwh.CompactMessage) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendCompactMessage", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendCompactMessage indicates an expected call of SendCompactMessage
func (mr *MockClientMockRecorder) SendCompactMessage(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendCompactMessage", reflect.TypeOf((*MockClient)(nil).SendCompactMessage), arg0, arg1)
}

// SendMessage mocks base method
func (m *MockClient) SendMessage(arg0 context.Context, arg1 dwh.Message) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendMessage", arg0, arg1)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendMessage indicates an expected call of SendMessage
func (mr *MockClientMockRecorder) SendMessage(arg0, arg1 interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendMessage", reflect.TypeOf((*MockClient)(nil).SendMessage), arg0, arg1)
}
