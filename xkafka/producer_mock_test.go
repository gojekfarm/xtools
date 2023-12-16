// Code generated by mockery v2.20.0. DO NOT EDIT.

package xkafka

import (
	kafka "github.com/confluentinc/confluent-kafka-go/v2/kafka"
	mock "github.com/stretchr/testify/mock"
)

// MockProducerClient is an autogenerated mock type for the producerClient type
type MockProducerClient struct {
	mock.Mock
}

// Close provides a mock function with given fields:
func (_m *MockProducerClient) Close() {
	_m.Called()
}

// Events provides a mock function with given fields:
func (_m *MockProducerClient) Events() chan kafka.Event {
	ret := _m.Called()

	var r0 chan kafka.Event
	if rf, ok := ret.Get(0).(func() chan kafka.Event); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(chan kafka.Event)
		}
	}

	return r0
}

// Flush provides a mock function with given fields: timeoutMs
func (_m *MockProducerClient) Flush(timeoutMs int) int {
	ret := _m.Called(timeoutMs)

	var r0 int
	if rf, ok := ret.Get(0).(func(int) int); ok {
		r0 = rf(timeoutMs)
	} else {
		r0 = ret.Get(0).(int)
	}

	return r0
}

// Produce provides a mock function with given fields: msg, deliveryChan
func (_m *MockProducerClient) Produce(msg *kafka.Message, deliveryChan chan kafka.Event) error {
	ret := _m.Called(msg, deliveryChan)

	var r0 error
	if rf, ok := ret.Get(0).(func(*kafka.Message, chan kafka.Event) error); ok {
		r0 = rf(msg, deliveryChan)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ProduceChannel provides a mock function with given fields:
func (_m *MockProducerClient) ProduceChannel() chan *kafka.Message {
	ret := _m.Called()

	var r0 chan *kafka.Message
	if rf, ok := ret.Get(0).(func() chan *kafka.Message); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(chan *kafka.Message)
		}
	}

	return r0
}

type mockConstructorTestingTNewMockProducerClient interface {
	mock.TestingT
	Cleanup(func())
}

// NewMockProducerClient creates a new instance of MockProducerClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockProducerClient(t mockConstructorTestingTNewMockProducerClient) *MockProducerClient {
	mock := &MockProducerClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
