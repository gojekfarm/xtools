// Code generated by mockery v2.20.0. DO NOT EDIT.

package xkafka_test

import (
	kafka "github.com/confluentinc/confluent-kafka-go/kafka"
	mock "github.com/stretchr/testify/mock"

	time "time"
)

// MockKafkaConsumer is an autogenerated mock type for the KafkaConsumer type
type MockKafkaConsumer struct {
	mock.Mock
}

// Close provides a mock function with given fields:
func (_m *MockKafkaConsumer) Close() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetMetadata provides a mock function with given fields: topic, allTopics, timeoutMs
func (_m *MockKafkaConsumer) GetMetadata(topic *string, allTopics bool, timeoutMs int) (*kafka.Metadata, error) {
	ret := _m.Called(topic, allTopics, timeoutMs)

	var r0 *kafka.Metadata
	var r1 error
	if rf, ok := ret.Get(0).(func(*string, bool, int) (*kafka.Metadata, error)); ok {
		return rf(topic, allTopics, timeoutMs)
	}
	if rf, ok := ret.Get(0).(func(*string, bool, int) *kafka.Metadata); ok {
		r0 = rf(topic, allTopics, timeoutMs)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*kafka.Metadata)
		}
	}

	if rf, ok := ret.Get(1).(func(*string, bool, int) error); ok {
		r1 = rf(topic, allTopics, timeoutMs)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ReadMessage provides a mock function with given fields: timeout
func (_m *MockKafkaConsumer) ReadMessage(timeout time.Duration) (*kafka.Message, error) {
	ret := _m.Called(timeout)

	var r0 *kafka.Message
	var r1 error
	if rf, ok := ret.Get(0).(func(time.Duration) (*kafka.Message, error)); ok {
		return rf(timeout)
	}
	if rf, ok := ret.Get(0).(func(time.Duration) *kafka.Message); ok {
		r0 = rf(timeout)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*kafka.Message)
		}
	}

	if rf, ok := ret.Get(1).(func(time.Duration) error); ok {
		r1 = rf(timeout)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SubscribeTopics provides a mock function with given fields: topics, rebalanceCb
func (_m *MockKafkaConsumer) SubscribeTopics(topics []string, rebalanceCb kafka.RebalanceCb) error {
	ret := _m.Called(topics, rebalanceCb)

	var r0 error
	if rf, ok := ret.Get(0).(func([]string, kafka.RebalanceCb) error); ok {
		r0 = rf(topics, rebalanceCb)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Unsubscribe provides a mock function with given fields:
func (_m *MockKafkaConsumer) Unsubscribe() error {
	ret := _m.Called()

	var r0 error
	if rf, ok := ret.Get(0).(func() error); ok {
		r0 = rf()
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewMockKafkaConsumer interface {
	mock.TestingT
	Cleanup(func())
}

// NewMockKafkaConsumer creates a new instance of MockKafkaConsumer. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewMockKafkaConsumer(t mockConstructorTestingTNewMockKafkaConsumer) *MockKafkaConsumer {
	mock := &MockKafkaConsumer{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}