package xkafka

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewBatchConsumer(t *testing.T) {
	t.Parallel()

	cfg := ConfigMap{
		"enable.auto.commit": true,
		"auto.offset.reset":  "earliest",
	}

	consumer, err := NewBatchConsumer(
		"test-batch-consumer",
		noopBatchHandler(),
		testTopics,
		testBrokers,
		cfg,
		PollTimeout(testTimeout),
		MetadataTimeout(testTimeout),
		ShutdownTimeout(testTimeout),
		Concurrency(2),
		ErrorHandler(NoopErrorHandler),
		BatchSize(10),
		BatchTimeout(testTimeout),
		ManualCommit(true),
	)
	assert.NoError(t, err)
	assert.NotNil(t, consumer)

	assert.Equal(t, "test-batch-consumer", consumer.name)
	assert.EqualValues(t, testTopics, consumer.config.topics)
	assert.EqualValues(t, testBrokers, consumer.config.brokers)
	assert.EqualValues(t, testTimeout, consumer.config.pollTimeout)
	assert.EqualValues(t, testTimeout, consumer.config.metadataTimeout)
	assert.EqualValues(t, testTimeout, consumer.config.shutdownTimeout)
	assert.EqualValues(t, 2, consumer.config.concurrency)
	assert.NotNil(t, consumer.config.errorHandler)
	assert.EqualValues(t, 10, consumer.config.batchSize)
	assert.EqualValues(t, testTimeout, consumer.config.batchTimeout)

	expectedConfig := kafka.ConfigMap{
		"bootstrap.servers":        "localhost:9092",
		"group.id":                 "test-batch-consumer",
		"auto.offset.reset":        "earliest",
		"enable.auto.commit":       false,
		"enable.auto.offset.store": false,
	}

	assert.EqualValues(t, expectedConfig, consumer.config.configMap)
}

func TestNewBatchConsumer_ConfigValidation(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name    string
		options []ConsumerOption
		expect  error
	}{
		{
			name:    "missing brokers",
			options: []ConsumerOption{testTopics, errHandler},
			expect:  ErrRequiredOption,
		},
		{
			name:    "missing topics",
			options: []ConsumerOption{testBrokers, errHandler},
			expect:  ErrRequiredOption,
		},
		{
			name:    "missing error handler",
			options: []ConsumerOption{testBrokers, testTopics},
			expect:  ErrRequiredOption,
		},
		{
			name: "consumer error",
			options: []ConsumerOption{
				testTopics, testBrokers, errHandler,
				consumerFunc(func(configMap *kafka.ConfigMap) (consumerClient, error) {
					return nil, assert.AnError
				}),
			},
			expect: assert.AnError,
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := NewBatchConsumer("test-batch-consumer", noopBatchHandler(), tc.options...)
			assert.Error(t, err)
			assert.ErrorIs(t, err, tc.expect)
		})
	}
}

func TestBatchConsumer_Lifecycle(t *testing.T) {
	t.Parallel()

	t.Run("RunSubscribeError", func(t *testing.T) {
		consumer, mockKafka := newTestBatchConsumer(t, defaultOpts...)

		expectError := errors.New("error in subscribe")

		mockKafka.On("SubscribeTopics", []string(testTopics), mock.Anything).Return(expectError)

		assert.Error(t, consumer.Run(context.Background()))

		mockKafka.AssertExpectations(t)
	})

	t.Run("RunUnsubscribeError", func(t *testing.T) {
		consumer, mockKafka := newTestBatchConsumer(t, defaultOpts...)

		mockKafka.On("SubscribeTopics", []string(testTopics), mock.Anything).Return(nil)
		mockKafka.On("ReadMessage", testTimeout).Return(newFakeKafkaMessage(), nil)
		mockKafka.On("Commit").Return(nil, nil)
		mockKafka.On("Close").Return(nil)

		mockKafka.On("Unsubscribe").Return(assert.AnError)

		ctx, cancel := context.WithCancel(context.Background())

		go func() {
			<-time.After(100 * time.Millisecond)
			cancel()
		}()

		assert.Error(t, consumer.Run(ctx))

		mockKafka.AssertExpectations(t)
	})

	t.Run("RunCloseError", func(t *testing.T) {
		consumer, mockKafka := newTestBatchConsumer(t, defaultOpts...)

		mockKafka.On("SubscribeTopics", []string(testTopics), mock.Anything).Return(nil)
		mockKafka.On("Unsubscribe").Return(nil)
		mockKafka.On("ReadMessage", testTimeout).Return(newFakeKafkaMessage(), nil)
		mockKafka.On("Commit").Return(nil, nil)
		mockKafka.On("Close").Return(errors.New("error in close"))

		ctx, cancel := context.WithCancel(context.Background())

		go func() {
			<-time.After(100 * time.Millisecond)
			cancel()
		}()

		assert.Error(t, consumer.Run(ctx))

		mockKafka.AssertExpectations(t)
	})

	t.Run("RunSuccess", func(t *testing.T) {
		consumer, mockKafka := newTestBatchConsumer(t, defaultOpts...)

		mockKafka.On("SubscribeTopics", []string(testTopics), mock.Anything).Return(nil)
		mockKafka.On("Unsubscribe").Return(nil)
		mockKafka.On("ReadMessage", testTimeout).Return(newFakeKafkaMessage(), nil)
		mockKafka.On("Commit").Return(nil, nil)
		mockKafka.On("Close").Return(nil)

		ctx, cancel := context.WithCancel(context.Background())

		go func() {
			<-time.After(100 * time.Millisecond)
			cancel()
		}()

		assert.NoError(t, consumer.Run(ctx))

		mockKafka.AssertExpectations(t)
	})
}

func TestBatchConsumer_HandleBatch(t *testing.T) {
	t.Parallel()

	opts := append(defaultOpts, BatchSize(10))
	consumer, mockKafka := newTestBatchConsumer(t, opts...)

	km := newFakeKafkaMessage()
	ctx, cancel := context.WithCancel(context.Background())

	handler := BatchHandlerFunc(func(ctx context.Context, b *Batch) error {
		assert.NotNil(t, b)
		assert.Len(t, b.Messages, 10)

		cancel()
		return nil
	})

	mockKafka.On("SubscribeTopics", []string(testTopics), mock.Anything).Return(nil)
	mockKafka.On("Unsubscribe").Return(nil)
	mockKafka.On("Commit").Return(nil, nil)
	mockKafka.On("ReadMessage", testTimeout).Return(km, nil)
	mockKafka.On("Close").Return(nil)

	consumer.handler = handler
	err := consumer.Run(ctx)
	assert.NoError(t, err)

	mockKafka.AssertExpectations(t)
}

func TestBatchConsumer_HandleBatchError(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name    string
		options []ConsumerOption
	}{
		{
			name: "sequential",
			options: []ConsumerOption{
				BatchSize(10),
				BatchTimeout(testTimeout),
			},
		},
		{
			name: "async",
			options: []ConsumerOption{
				Concurrency(2),
				BatchSize(10),
				BatchTimeout(testTimeout),
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			consumer, mockKafka := newTestBatchConsumer(t, append(defaultOpts, tc.options...)...)

			km := newFakeKafkaMessage()
			ctx := context.Background()

			handler := BatchHandlerFunc(func(ctx context.Context, b *Batch) error {
				err := assert.AnError

				return b.AckFail(err)
			})

			mockKafka.On("SubscribeTopics", []string(testTopics), mock.Anything).Return(nil)
			mockKafka.On("Unsubscribe").Return(nil)
			mockKafka.On("Commit").Return(nil, nil)
			mockKafka.On("ReadMessage", testTimeout).Return(km, nil)
			mockKafka.On("Close").Return(nil)

			consumer.handler = handler
			err := consumer.Run(ctx)
			assert.ErrorIs(t, err, assert.AnError)

			mockKafka.AssertExpectations(t)
		})
	}
}

func TestBatchConsumer_Async(t *testing.T) {
	t.Parallel()

	opts := append(defaultOpts,
		Concurrency(2),
		BatchSize(3),
	)
	consumer, mockKafka := newTestBatchConsumer(t, opts...)

	km := newFakeKafkaMessage()
	ctx, cancel := context.WithCancel(context.Background())

	count := atomic.Int32{}

	handler := BatchHandlerFunc(func(ctx context.Context, b *Batch) error {
		b.AckSuccess()

		assert.NotNil(t, b)

		n := count.Add(1)

		if n == 2 {
			cancel()
		}

		return nil
	})

	mockKafka.On("SubscribeTopics", []string(testTopics), mock.Anything).Return(nil)
	mockKafka.On("Unsubscribe").Return(nil)
	mockKafka.On("Commit").Return(nil, nil)
	mockKafka.On("StoreOffsets", mock.Anything).Return(nil, nil)
	mockKafka.On("ReadMessage", testTimeout).Return(km, nil)
	mockKafka.On("Close").Return(nil)

	consumer.handler = handler
	err := consumer.Run(ctx)
	assert.NoError(t, err)

	mockKafka.AssertExpectations(t)
}

func TestBatchConsumer_StopOffsetOnError(t *testing.T) {
	t.Parallel()

	opts := append(defaultOpts,
		Concurrency(2),
		BatchSize(3),
	)
	consumer, mockKafka := newTestBatchConsumer(t, opts...)

	km := newFakeKafkaMessage()
	ctx, cancel := context.WithCancel(context.Background())

	count := atomic.Int32{}

	handler := BatchHandlerFunc(func(ctx context.Context, b *Batch) error {
		assert.NotNil(t, b)

		n := count.Add(1)

		if n > 2 {
			err := assert.AnError
			cancel()

			return b.AckFail(err)
		}

		b.AckSuccess()

		return nil
	})

	mockKafka.On("SubscribeTopics", []string(testTopics), mock.Anything).Return(nil)
	mockKafka.On("Unsubscribe").Return(nil)
	mockKafka.On("Commit").Return(nil, nil)
	mockKafka.On("ReadMessage", testTimeout).Return(km, nil)
	mockKafka.On("Close").Return(nil)

	mockKafka.On("StoreOffsets", mock.Anything).
		Return(nil, nil).
		Times(2)

	consumer.handler = handler
	err := consumer.Run(ctx)
	assert.ErrorIs(t, err, assert.AnError)

	mockKafka.AssertExpectations(t)
}

func TestBatchConsumer_BatchTimeout(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name    string
		options []ConsumerOption
	}{
		{
			name: "sequential",
			options: []ConsumerOption{
				BatchTimeout(10 * time.Millisecond),
				BatchSize(100_000),
			},
		},
		{
			name: "async",
			options: []ConsumerOption{
				Concurrency(2),
				BatchTimeout(10 * time.Millisecond),
				BatchSize(100_000),
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			consumer, mockKafka := newTestBatchConsumer(t, append(defaultOpts, tc.options...)...)

			km := newFakeKafkaMessage()
			ctx, cancel := context.WithCancel(context.Background())

			handler := BatchHandlerFunc(func(ctx context.Context, b *Batch) error {
				b.AckSuccess()

				assert.NotNil(t, b)
				assert.True(t, len(b.Messages) > 0)

				return nil
			})

			mockKafka.On("SubscribeTopics", []string(testTopics), mock.Anything).Return(nil)
			mockKafka.On("Unsubscribe").Return(nil)
			mockKafka.On("Commit").Return(nil, nil)
			mockKafka.On("StoreOffsets", mock.Anything).Return(nil, nil)
			mockKafka.On("ReadMessage", testTimeout).Return(km, nil)
			mockKafka.On("Close").Return(nil)

			consumer.handler = handler

			go func() {
				<-time.After(500 * time.Millisecond)
				cancel()
			}()

			err := consumer.Run(ctx)
			assert.NoError(t, err)

			mockKafka.AssertExpectations(t)
		})
	}
}

func TestBatchConsumer_MiddlewareExecutionOrder(t *testing.T) {
	t.Parallel()

	consumer, mockKafka := newTestBatchConsumer(t, defaultOpts...)

	km := newFakeKafkaMessage()
	ctx, cancel := context.WithCancel(context.Background())

	mockKafka.On("SubscribeTopics", []string(testTopics), mock.Anything).Return(nil)
	mockKafka.On("Unsubscribe").Return(nil)
	mockKafka.On("Commit").Return(nil, nil)
	mockKafka.On("ReadMessage", testTimeout).Return(km, nil)
	mockKafka.On("Close").Return(nil)

	handler := BatchHandlerFunc(func(ctx context.Context, b *Batch) error {
		cancel()

		return nil
	})

	var preExec []string
	var postExec []string

	middlewares := []BatchMiddlewarer{
		testBatchMiddleware("middleware1", &preExec, &postExec),
		testBatchMiddleware("middleware2", &preExec, &postExec),
		testBatchMiddleware("middleware3", &preExec, &postExec),
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

func TestBatchConsumer_ManualCommit(t *testing.T) {
	t.Parallel()

	consumer, mockKafka := newTestBatchConsumer(t, defaultOpts...)

	km := newFakeKafkaMessage()
	ctx, cancel := context.WithCancel(context.Background())

	mockKafka.On("SubscribeTopics", []string(testTopics), mock.Anything).Return(nil)
	mockKafka.On("Unsubscribe").Return(nil)
	mockKafka.On("StoreOffsets", mock.Anything).Return(nil, nil)
	mockKafka.On("Commit").Return(nil, nil)
	mockKafka.On("ReadMessage", testTimeout).Return(km, nil)
	mockKafka.On("Close").Return(nil)

	handler := BatchHandlerFunc(func(ctx context.Context, b *Batch) error {
		b.AckSuccess()

		cancel()

		return nil
	})

	consumer.handler = handler
	err := consumer.Run(ctx)
	assert.NoError(t, err)

	mockKafka.AssertExpectations(t)
}

func TestBatchConsumer_ReadMessageTimeout(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name    string
		options []ConsumerOption
	}{
		{
			name: "sequential",
			options: []ConsumerOption{
				BatchSize(2),
			},
		},
		{
			name: "async",
			options: []ConsumerOption{
				Concurrency(2),
				BatchSize(2),
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			consumer, mockKafka := newTestBatchConsumer(t, append(defaultOpts, tc.options...)...)

			ctx, cancel := context.WithCancel(context.Background())
			expect := kafka.NewError(kafka.ErrTimedOut, "kafka: timed out", false)
			km := newFakeKafkaMessage()

			counter := atomic.Int32{}

			handler := BatchHandlerFunc(func(ctx context.Context, b *Batch) error {
				n := counter.Add(1)

				if n == 1 {
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
			mockKafka.On("Close").Return(nil)

			consumer.handler = handler

			err := consumer.Run(ctx)
			assert.NoError(t, err)

			mockKafka.AssertExpectations(t)
		})
	}
}

func TestBatchConsumer_KafkaError(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name    string
		options []ConsumerOption
	}{
		{
			name: "sequential",
			options: []ConsumerOption{
				BatchSize(2),
			},
		},
		{
			name: "async",
			options: []ConsumerOption{
				Concurrency(2),
				BatchSize(2),
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			consumer, mockKafka := newTestBatchConsumer(t, append(defaultOpts, tc.options...)...)

			ctx := context.Background()
			expect := kafka.NewError(kafka.ErrUnknown, "kafka: unknown error", false)
			km := newFakeKafkaMessage()

			mockKafka.On("SubscribeTopics", []string(testTopics), mock.Anything).Return(nil)
			mockKafka.On("Unsubscribe").Return(nil)
			mockKafka.On("Commit").Return(nil, nil)
			mockKafka.On("Close").Return(nil)

			mockKafka.On("ReadMessage", testTimeout).
				Return(km, nil).
				Times(3)
			mockKafka.On("ReadMessage", testTimeout).
				Return(nil, expect).
				Once()

			err := consumer.Run(ctx)
			assert.Error(t, err)
			assert.ErrorIs(t, err, expect)

			mockKafka.AssertExpectations(t)
		})
	}
}

func TestBatchConsumer_CommitError(t *testing.T) {
	t.Parallel()

	testcases := []struct {
		name    string
		options []ConsumerOption
	}{
		{
			name: "sequential",
			options: []ConsumerOption{
				ManualCommit(true),
			},
		},
		{
			name: "async",
			options: []ConsumerOption{
				ManualCommit(true),
				Concurrency(2),
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			consumer, mockKafka := newTestBatchConsumer(t, append(defaultOpts, tc.options...)...)

			km := newFakeKafkaMessage()
			ctx := context.Background()
			expect := errors.New("error in commit")

			handler := BatchHandlerFunc(func(ctx context.Context, b *Batch) error {
				b.AckSuccess()

				return nil
			})

			mockKafka.On("SubscribeTopics", []string(testTopics), mock.Anything).Return(nil)
			mockKafka.On("Unsubscribe").Return(nil)
			mockKafka.On("StoreOffsets", mock.Anything).Return(nil, nil)
			mockKafka.On("ReadMessage", testTimeout).Return(km, nil)
			mockKafka.On("Close").Return(nil)

			mockKafka.On("Commit").Return(nil, expect)

			consumer.handler = handler

			err := consumer.Run(ctx)
			assert.Error(t, err)
			assert.ErrorIs(t, err, expect)

			mockKafka.AssertExpectations(t)
		})
	}
}

func noopBatchHandler() BatchHandler {
	return BatchHandlerFunc(func(ctx context.Context, b *Batch) error {
		return nil
	})
}

func newTestBatchConsumer(t *testing.T, opts ...ConsumerOption) (*BatchConsumer, *MockConsumerClient) {
	t.Helper()

	mockConsumer := &MockConsumerClient{}

	opts = append(opts, mockConsumerFunc(mockConsumer))

	consumer, err := NewBatchConsumer(
		"test-batch-consumer",
		noopBatchHandler(),
		opts...,
	)
	require.NoError(t, err)
	require.NotNil(t, consumer)

	return consumer, mockConsumer
}

func testBatchMiddleware(name string, preExec, postExec *[]string) BatchMiddlewarer {
	return BatchMiddlewareFunc(func(next BatchHandler) BatchHandler {
		return BatchHandlerFunc(func(ctx context.Context, b *Batch) error {
			*preExec = append(*preExec, name)
			err := next.HandleBatch(ctx, b)
			*postExec = append(*postExec, name)

			return err
		})
	})
}
