package xkafka

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	testTopics      = Topics{"test-topic"}
	testBrokers     = Brokers{"localhost:9092"}
	testPollTimeout = 1 * time.Second
)

func TestConsumerGetMetadata(t *testing.T) {
	consumer, mockKafka := newTestConsumer(t)

	mockKafka.On("GetMetadata", mock.Anything, false, 10000).Return(&kafka.Metadata{}, nil)

	metadata, err := consumer.GetMetadata()
	assert.NoError(t, err)
	assert.NotNil(t, metadata)

	consumer.Close()

	mockKafka.AssertExpectations(t)
}

func TestConsumerHandleMessage(t *testing.T) {
	consumer, mockKafka := newTestConsumer(t)

	km := newFakeKafkaMessage()
	ctx, cancel := context.WithCancel(context.Background())

	handler := HandlerFunc(func(ctx context.Context, msg *Message) error {
		assert.Equal(t, km.Key, msg.Key)
		assert.Equal(t, km.Value, msg.Value)

		cancel()
		return nil
	})

	mockKafka.On("SubscribeTopics", []string(testTopics), mock.Anything).Return(nil)
	mockKafka.On("Unsubscribe").Return(nil)
	mockKafka.On("ReadMessage", testPollTimeout).Return(km, nil)

	consumer.handler = handler
	err := consumer.Run(ctx)
	assert.NoError(t, err)

	mockKafka.AssertExpectations(t)
}

func TestConsumerHandleMessageError(t *testing.T) {
	consumer, mockKafka := newTestConsumer(t)

	km := newFakeKafkaMessage()
	ctx := context.Background()
	expect := errors.New("error in handler")

	handler := HandlerFunc(func(ctx context.Context, msg *Message) error {
		return expect
	})

	mockKafka.On("SubscribeTopics", []string(testTopics), mock.Anything).Return(nil)
	mockKafka.On("Unsubscribe").Return(nil)
	mockKafka.On("ReadMessage", testPollTimeout).Return(km, nil)

	consumer.handler = handler
	err := consumer.Run(ctx)
	assert.Error(t, err)
	assert.Equal(t, expect, err)

	mockKafka.AssertExpectations(t)
}

func TestConsumerReadMessageTimeout(t *testing.T) {
	consumer, mockKafka := newTestConsumer(t)

	ctx, cancel := context.WithCancel(context.Background())
	expect := kafka.NewError(kafka.ErrTimedOut, "kafka: timed out", false)
	km := newFakeKafkaMessage()

	counter := 0
	handler := HandlerFunc(func(ctx context.Context, msg *Message) error {
		counter++

		if counter > 2 {
			cancel()
		}

		return nil
	})

	mockKafka.On("SubscribeTopics", []string(testTopics), mock.Anything).Return(nil)
	mockKafka.On("Unsubscribe").Return(nil)
	mockKafka.On("ReadMessage", testPollTimeout).Return(km, nil).Once()
	mockKafka.On("ReadMessage", testPollTimeout).Return(nil, expect).Once()
	mockKafka.On("ReadMessage", testPollTimeout).Return(km, nil)

	consumer.handler = handler

	err := consumer.Run(ctx)
	assert.NoError(t, err)

	mockKafka.AssertExpectations(t)
}

func TestConsumerKafkaError(t *testing.T) {
	consumer, mockKafka := newTestConsumer(t)

	ctx := context.Background()
	expect := kafka.NewError(kafka.ErrUnknown, "kafka: unknown error", false)
	km := newFakeKafkaMessage()

	mockKafka.On("SubscribeTopics", []string(testTopics), mock.Anything).Return(nil)
	mockKafka.On("Unsubscribe").Return(nil)
	mockKafka.On("ReadMessage", testPollTimeout).Return(km, nil).Once()
	mockKafka.On("ReadMessage", testPollTimeout).Return(nil, expect).Once()

	err := consumer.Run(ctx)
	assert.Error(t, err)
	assert.Equal(t, expect, err)

	mockKafka.AssertExpectations(t)
}

func TestConsumerMiddlewareExecutionOrder(t *testing.T) {
	consumer, mockKafka := newTestConsumer(t)

	km := newFakeKafkaMessage()
	ctx, cancel := context.WithCancel(context.Background())

	mockKafka.On("SubscribeTopics", []string(testTopics), mock.Anything).Return(nil)
	mockKafka.On("Unsubscribe").Return(nil)
	mockKafka.On("ReadMessage", testPollTimeout).Return(km, nil).Once()

	handler := HandlerFunc(func(ctx context.Context, msg *Message) error {
		cancel()

		return nil
	})

	preExec := []string{}
	postExec := []string{}

	middlewares := []MiddlewareFunc{
		testMiddleware("middleware1", &preExec, &postExec),
		testMiddleware("middleware2", &preExec, &postExec),
		testMiddleware("middleware3", &preExec, &postExec),
	}

	consumer.Use(middlewares...)
	consumer.handler = handler

	err := consumer.Run(ctx)
	assert.NoError(t, err)
	assert.Equal(t, []string{"middleware1", "middleware2", "middleware3"}, preExec)
	assert.Equal(t, []string{"middleware3", "middleware2", "middleware1"}, postExec)

	mockKafka.AssertExpectations(t)
}

func testMiddleware(name string, pre, post *[]string) MiddlewareFunc {
	return func(next Handler) Handler {
		return HandlerFunc(func(ctx context.Context, msg *Message) error {
			*pre = append(*pre, name)
			err := next.Handle(ctx, msg)
			*post = append(*post, name)

			return err
		})
	}
}

func newTestConsumer(t *testing.T) (*Consumer, *MockConsumerClient) {
	mockConsumer := &MockConsumerClient{}

	mockConsumer.On("Close").Return(nil)

	opts := []Option{
		testTopics,
		testBrokers,
		mockConsumerFunc(mockConsumer),
		PollTimeout(testPollTimeout),
	}
	consumer, err := NewConsumer("consumer-id", noopHandler(), opts...)
	require.NoError(t, err)
	require.NotNil(t, consumer)

	return consumer, mockConsumer
}

func newFakeKafkaMessage() *kafka.Message {
	return &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &testTopics[0],
			Partition: 1,
		},
		Key:       []byte("key"),
		Value:     []byte("value"),
		Timestamp: time.Date(2020, 1, 1, 23, 59, 59, 0, time.UTC),
	}
}

func mockConsumerFunc(mock *MockConsumerClient) consumerFunc {
	return func(configMap *kafka.ConfigMap) (consumerClient, error) {
		return mock, nil
	}
}

func noopHandler() Handler {
	return HandlerFunc(func(ctx context.Context, msg *Message) error {
		return nil
	})
}
