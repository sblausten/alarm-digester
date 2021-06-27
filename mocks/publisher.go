// Code generated by MockGen. DO NOT EDIT.
// Source: src/nats/publisher.go

// Package nats is a generated GoMock package.
package mocks

import (
	reflect "reflect"

	gomock "github.com/golang/mock/gomock"
	models "github.com/sblausten/go-service/models"
)

// MockPublisherInterface is a mock of PublisherInterface interface.
type MockPublisherInterface struct {
	ctrl     *gomock.Controller
	recorder *MockPublisherInterfaceMockRecorder
}

// MockPublisherInterfaceMockRecorder is the mock recorder for MockPublisherInterface.
type MockPublisherInterfaceMockRecorder struct {
	mock *MockPublisherInterface
}

// NewMockPublisherInterface creates a new mock instance.
func NewMockPublisherInterface(ctrl *gomock.Controller) *MockPublisherInterface {
	mock := &MockPublisherInterface{ctrl: ctrl}
	mock.recorder = &MockPublisherInterfaceMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockPublisherInterface) EXPECT() *MockPublisherInterfaceMockRecorder {
	return m.recorder
}

// PublishMessage mocks base method.
func (m *MockPublisherInterface) PublishMessage(subject string, message models.AlarmDigest) {
	m.ctrl.T.Helper()
	m.ctrl.Call(m, "PublishMessage", subject, message)
}

// PublishMessage indicates an expected call of PublishMessage.
func (mr *MockPublisherInterfaceMockRecorder) PublishMessage(subject, message interface{}) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "PublishMessage", reflect.TypeOf((*MockPublisherInterface)(nil).PublishMessage), subject, message)
}