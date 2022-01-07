// Code generated by MockGen. DO NOT EDIT.
// Source: ../telemetry/telemetry.go

// Package core is a generated GoMock package.
package core

import (
	reflect "reflect"

	telemetry "github.com/ChainSafe/gossamer/dot/telemetry"
	gomock "github.com/golang/mock/gomock"
)

// MockClient is a mock of Client interface.
type MockClient struct {
	ctrl     *gomock.Controller
	recorder *MockClientMockRecorder
}

// MockClientMockRecorder is the mock recorder for MockClient.
type MockClientMockRecorder struct {
	mock *MockClient
}

// NewMockClient creates a new mock instance.
func NewMockClient(ctrl *gomock.Controller) *MockClient {
	mock := &MockClient{ctrl: ctrl}
	mock.recorder = &MockClientMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockClient) EXPECT() *MockClientMockRecorder {
	return m.recorder
}

// SendMessage mocks base method.
func (m *MockClient) SendMessage(msg telemetry.Message) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "SendMessage", msg)
	ret0, _ := ret[0].(error)
	return ret0
}

// SendMessage indicates an expected call of SendMessage.
func (mr *MockClientMockRecorder) SendMessage(msg interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "SendMessage", reflect.TypeOf((*MockClient)(nil).SendMessage), msg)
}

// MockMessage is a mock of Message interface.
type MockMessage struct {
	ctrl     *gomock.Controller
	recorder *MockMessageMockRecorder
}

// MockMessageMockRecorder is the mock recorder for MockMessage.
type MockMessageMockRecorder struct {
	mock *MockMessage
}

// NewMockMessage creates a new mock instance.
func NewMockMessage(ctrl *gomock.Controller) *MockMessage {
	mock := &MockMessage{ctrl: ctrl}
	mock.recorder = &MockMessageMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockMessage) EXPECT() *MockMessageMockRecorder {
	return m.recorder
}

// messageType mocks base method.
func (m *MockMessage) messageType() string {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "messageType")
	ret0, _ := ret[0].(string)
	return ret0
}

// messageType indicates an expected call of messageType.
func (mr *MockMessageMockRecorder) messageType() *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "messageType", reflect.TypeOf((*MockMessage)(nil).messageType))
}
