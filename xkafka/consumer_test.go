package xkafka

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	testTopics  = Topics{"test-topic"}
	testBrokers = Brokers{"localhost:9092"}
	testTimeout = time.Second
	defaultOpts = []Option{
		testTopics,
		testBrokers,
		PollTimeout(testTimeout),
	}
)

func TestNewConsumer(t *testing.T) {
	t.Parallel()

	cfg := ConfigMap{
		"enable.auto.commit": true,
		"auto.offset.reset":  "earliest",
	}

	consumer, err := NewConsumer(
		"test-consumer",
		noopHandler(),
		testTopics,
		testBrokers,
		cfg,
		PollTimeout(testTimeout),
		MetadataTimeout(testTimeout),
		ShutdownTimeout(testTimeout),
		Concurrency(2),
		ErrorHandler(NoopErrorHandler),
	)
	assert.NoError(t, err)
	assert.NotNil(t, consumer)

	assert.Equal(t, "test-consumer", consumer.name)
	assert.EqualValues(t, testTopics, consumer.config.topics)
	assert.EqualValues(t, testBrokers, consumer.config.brokers)
	assert.EqualValues(t, testTimeout, consumer.config.pollTimeout)
	assert.EqualValues(t, testTimeout, consumer.config.metadataTimeout)
	assert.EqualValues(t, testTimeout, consumer.config.shutdownTimeout)
	assert.EqualValues(t, 2, consumer.config.concurrency)
	assert.NotNil(t, consumer.config.errorHandler)

	expectedConfig := kafka.ConfigMap{
		"bootstrap.servers":        "localhost:9092",
		"group.id":                 "test-consumer",
		"auto.offset.reset":        "earliest",
		"enable.auto.commit":       true,
		"enable.auto.offset.store": false,
	}

	assert.EqualValues(t, expectedConfig, consumer.config.configMap)
}

func TestNewConsumerError(t *testing.T) {
	t.Parallel()

	expectError := errors.New("error in consumer")

	fn := func(configMap *kafka.ConfigMap) (consumerClient, error) {
		return nil, expectError
	}

	_, err := NewConsumer(
		"test-consumer",
		noopHandler(),
		consumerFunc(fn),
	)
	assert.Error(t, err)
	assert.Equal(t, expectError, err)
}

func TestConsumerGetMetadata(t *testing.T) {
	t.Parallel()

	consumer, mockKafka := newTestConsumer(t,
		testTopics,
		testBrokers,
		PollTimeout(testTimeout),
	)

	mockKafka.On("GetMetadata", mock.Anything, false, 10000).Return(&kafka.Metadata{}, nil)

	metadata, err := consumer.GetMetadata()
	assert.NoError(t, err)
	assert.NotNil(t, metadata)

	consumer.Close()

	mockKafka.AssertExpectations(t)
}

func TestConsumerSubscribeError(t *testing.T) {
	t.Parallel()

	consumer, mockKafka := newTestConsumer(t,
		testTopics,
		testBrokers,
		PollTimeout(testTimeout),
	)

	ctx := context.Background()
	expectError := errors.New("error")

	mockKafka.On("SubscribeTopics", []string(testTopics), mock.Anything).Return(expectError)

	err := consumer.Run(ctx)
	assert.EqualError(t, err, expectError.Error())

	mockKafka.AssertExpectations(t)
}

func TestConsumerUnsubscribeError(t *testing.T) {
	t.Parallel()

	consumer, mockKafka := newTestConsumer(t,
		testTopics,
		testBrokers,
		PollTimeout(testTimeout),
	)

	km := newFakeKafkaMessage()
	ctx := context.Background()
	unsubError := errors.New("error in unsubscribe")
	handlerError := errors.New("error in handler")

	handler := HandlerFunc(func(ctx context.Context, msg *Message) error {
		return handlerError
	})

	mockKafka.On("SubscribeTopics", []string(testTopics), mock.Anything).Return(nil)
	mockKafka.On("Unsubscribe").Return(unsubError)
	mockKafka.On("ReadMessage", testTimeout).Return(km, nil)
	mockKafka.On("Commit").Return(nil, nil)

	consumer.handler = handler

	err := consumer.Run(ctx)
	assert.ErrorIs(t, err, handlerError)

	mockKafka.AssertExpectations(t)
}

func TestConsumerHandleMessage(t *testing.T) {
	t.Parallel()

	consumer, mockKafka := newTestConsumer(t,
		testTopics,
		testBrokers,
		PollTimeout(testTimeout),
	)

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
	mockKafka.On("Commit").Return(nil, nil)
	mockKafka.On("ReadMessage", testTimeout).Return(km, nil)

	consumer.handler = handler
	err := consumer.Run(ctx)
	assert.NoError(t, err)

	mockKafka.AssertExpectations(t)
}

func TestConsumerHandleMessageError(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name    string
		options []Option
	}{
		{
			name:    "sequential",
			options: []Option{},
		},
		{
			name: "async",
			options: []Option{
				Concurrency(2),
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			consumer, mockKafka := newTestConsumer(t, append(defaultOpts, tc.options...)...)
			km := newFakeKafkaMessage()
			ctx := context.Background()
			expect := errors.New("error in handler")

			handler := HandlerFunc(func(ctx context.Context, msg *Message) error {
				msg.AckFail(expect)

				return expect
			})

			mockKafka.On("SubscribeTopics", []string(testTopics), mock.Anything).Return(nil)
			mockKafka.On("Unsubscribe").Return(nil)
			mockKafka.On("Commit").Return(nil, nil)
			mockKafka.On("ReadMessage", testTimeout).Return(km, nil)

			consumer.handler = handler
			err := consumer.Run(ctx)
			assert.Error(t, err)
			assert.ErrorIs(t, err, expect)

			mockKafka.AssertExpectations(t)
		})
	}
}

func TestConsumerErrorCallback(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name    string
		options []Option
	}{
		{
			name:    "sequential",
			options: []Option{},
		},
		{
			name: "async",
			options: []Option{
				Concurrency(2),
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			km := newFakeKafkaMessage()
			ctx := context.Background()
			expect := errors.New("error in handler")

			errHandler := ErrorHandler(func(err error) error {
				assert.Equal(t, expect, err)

				return err
			})

			opts := append(defaultOpts, errHandler)
			opts = append(opts, tc.options...)

			consumer, mockKafka := newTestConsumer(t, opts...)

			handler := HandlerFunc(func(ctx context.Context, msg *Message) error {
				return expect
			})

			mockKafka.On("SubscribeTopics", []string(testTopics), mock.Anything).Return(nil)
			mockKafka.On("Unsubscribe").Return(nil)
			mockKafka.On("Commit").Return(nil, nil)
			mockKafka.On("ReadMessage", testTimeout).Return(km, nil)

			consumer.handler = handler

			err := consumer.Run(ctx)
			assert.EqualError(t, err, expect.Error())

			mockKafka.AssertExpectations(t)
		})
	}
}

func TestConsumerReadMessageTimeout(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name    string
		options []Option
	}{
		{
			name:    "sequential",
			options: []Option{},
		},
		{
			name: "async",
			options: []Option{
				Concurrency(2),
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			consumer, mockKafka := newTestConsumer(t, append(defaultOpts, tc.options...)...)

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
			mockKafka.On("Commit").Return(nil, nil)
			mockKafka.On("ReadMessage", testTimeout).Return(km, nil).Once()
			mockKafka.On("ReadMessage", testTimeout).Return(nil, expect).Once()
			mockKafka.On("ReadMessage", testTimeout).Return(km, nil)

			consumer.handler = handler

			err := consumer.Run(ctx)
			assert.NoError(t, err)

			mockKafka.AssertExpectations(t)
		})
	}
}

func TestConsumerKafkaError(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name    string
		options []Option
	}{
		{
			name:    "sequential",
			options: []Option{},
		},
		{
			name: "async",
			options: []Option{
				Concurrency(2),
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			consumer, mockKafka := newTestConsumer(t, append(defaultOpts, tc.options...)...)

			ctx := context.Background()
			expect := kafka.NewError(kafka.ErrUnknown, "kafka: unknown error", false)
			km := newFakeKafkaMessage()

			mockKafka.On("SubscribeTopics", []string(testTopics), mock.Anything).Return(nil)
			mockKafka.On("Unsubscribe").Return(nil)
			mockKafka.On("Commit").Return(nil, nil)
			mockKafka.On("ReadMessage", testTimeout).Return(km, nil).Once()
			mockKafka.On("ReadMessage", testTimeout).Return(nil, expect).Once()

			err := consumer.Run(ctx)
			assert.Error(t, err)
			assert.ErrorIs(t, err, expect)

			mockKafka.AssertExpectations(t)
		})
	}
}

func TestConsumerMiddlewareExecutionOrder(t *testing.T) {
	t.Parallel()

	consumer, mockKafka := newTestConsumer(t,
		testTopics,
		testBrokers,
		PollTimeout(testTimeout),
	)

	km := newFakeKafkaMessage()
	ctx, cancel := context.WithCancel(context.Background())

	mockKafka.On("SubscribeTopics", []string(testTopics), mock.Anything).Return(nil)
	mockKafka.On("Unsubscribe").Return(nil)
	mockKafka.On("Commit").Return(nil, nil)
	mockKafka.On("ReadMessage", testTimeout).Return(km, nil)

	handler := HandlerFunc(func(ctx context.Context, msg *Message) error {
		cancel()

		return nil
	})

	var preExec []string
	var postExec []string

	middlewares := []MiddlewareFunc{
		testMiddleware("middleware1", &preExec, &postExec),
		testMiddleware("middleware2", &preExec, &postExec),
		testMiddleware("middleware3", &preExec, &postExec),
	}

	consumer.Use(middlewares...)
	consumer.handler = handler

	err := consumer.Run(ctx)
	assert.NoError(t, err)

	// middleware execution order should be FIFO
	// but we only test the first 3 values because the
	// consumer keeps reading messages until the context
	// is canceled
	assert.Equal(t, []string{"middleware1", "middleware2", "middleware3"}, preExec[:3])
	assert.Equal(t, []string{"middleware3", "middleware2", "middleware1"}, postExec[:3])

	mockKafka.AssertExpectations(t)
}

func TestConsumerManualCommit(t *testing.T) {
	t.Parallel()

	consumer, mockKafka := newTestConsumer(t,
		testTopics,
		testBrokers,
		PollTimeout(testTimeout),
		ManualCommit(true),
	)

	km := newFakeKafkaMessage()
	ctx, cancel := context.WithCancel(context.Background())

	mockKafka.On("SubscribeTopics", []string(testTopics), mock.Anything).Return(nil)
	mockKafka.On("Unsubscribe").Return(nil)
	mockKafka.On("StoreOffsets", mock.Anything).Return(nil, nil)
	mockKafka.On("Commit").Return(nil, nil)
	mockKafka.On("ReadMessage", testTimeout).Return(km, nil)

	handler := HandlerFunc(func(ctx context.Context, msg *Message) error {
		cancel()

		msg.AckSuccess()

		return nil
	})

	consumer.handler = handler

	err := consumer.Run(ctx)
	assert.NoError(t, err)

	mockKafka.AssertExpectations(t)
}

func TestConsumerAsync(t *testing.T) {
	t.Parallel()

	consumer, mockKafka := newTestConsumer(t,
		testTopics,
		testBrokers,
		PollTimeout(testTimeout),
		Concurrency(2),
	)

	km := newFakeKafkaMessage()
	ctx, cancel := context.WithCancel(context.Background())

	mockKafka.On("SubscribeTopics", []string(testTopics), mock.Anything).Return(nil)
	mockKafka.On("Unsubscribe").Return(nil)
	mockKafka.On("StoreOffsets", mock.Anything).Return(nil, nil)
	mockKafka.On("ReadMessage", testTimeout).Return(km, nil)
	mockKafka.On("Commit").Return(nil, nil)

	var recv []*Message
	var mu sync.Mutex

	handler := HandlerFunc(func(ctx context.Context, msg *Message) error {
		mu.Lock()
		defer mu.Unlock()

		recv = append(recv, msg)

		msg.AckSuccess()

		if len(recv) > 2 {
			cancel()
		}

		return nil
	})

	consumer.handler = handler

	err := consumer.Run(ctx)
	assert.NoError(t, err)

	mockKafka.AssertExpectations(t)
}

func TestConsumerStoreOffsetsError(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name    string
		options []Option
	}{
		// {
		// 	name:    "sequential",
		// 	options: []Option{},
		// },
		{
			name: "async",
			options: []Option{
				Concurrency(2),
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			consumer, mockKafka := newTestConsumer(t, append(defaultOpts, tc.options...)...)

			km := newFakeKafkaMessage()
			ctx := context.Background()
			expect := errors.New("error in store offsets")

			handler := HandlerFunc(func(ctx context.Context, msg *Message) error {
				time.Sleep(20 * time.Millisecond)

				msg.AckSuccess()

				return nil
			})

			mockKafka.On("SubscribeTopics", []string(testTopics), mock.Anything).Return(nil)
			mockKafka.On("Unsubscribe").Return(nil)
			mockKafka.On("StoreOffsets", mock.Anything).Return(nil, expect)
			mockKafka.On("Commit").Return(nil, nil)
			mockKafka.On("ReadMessage", testTimeout).Return(km, nil)

			consumer.handler = handler

			err := consumer.Run(ctx)
			assert.Error(t, err)
			assert.ErrorIs(t, err, expect)

			mockKafka.AssertExpectations(t)
		})
	}
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

func newTestConsumer(t *testing.T, opts ...Option) (*Consumer, *MockConsumerClient) {
	mockConsumer := &MockConsumerClient{}

	mockConsumer.On("Close").Return(nil)

	opts = append(opts, mockConsumerFunc(mockConsumer))

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
