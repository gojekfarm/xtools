package xkafka

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewProducer(t *testing.T) {
	cfg := ConfigMap{
		"socket.keepalive.enable": true,
	}

	producer, err := NewProducer(
		"test-producer",
		testBrokers,
		cfg,
		ShutdownTimeout(testTimeout),
	)
	assert.NoError(t, err)
	assert.NotNil(t, producer)

	assert.EqualValues(t, testBrokers, producer.config.brokers)
	assert.EqualValues(t, testTimeout, producer.config.shutdownTimeout)

	expectedConfig := kafka.ConfigMap{
		"socket.keepalive.enable": true,
		"bootstrap.servers":       "localhost:9092",
		"client.id":               "test-producer",
		"default.topic.config": kafka.ConfigMap{
			"acks":        1,
			"partitioner": "consistent_random",
		},
	}

	assert.EqualValues(t, expectedConfig, producer.config.configMap)
}

func TestNewProducerError(t *testing.T) {
	expectError := errors.New("error in producer")

	fn := func(configMap *kafka.ConfigMap) (producerClient, error) {
		return nil, expectError
	}

	producer, err := NewProducer(
		"test-producer",
		testBrokers,
		ConfigMap{},
		producerFunc(fn),
	)
	assert.EqualError(t, err, expectError.Error())
	assert.Nil(t, producer)
}

func TestProducerPublish(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	msg := newFakeMessage()
	km := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &msg.Topic,
			Partition: kafka.PartitionAny,
		},
		Key:           msg.Key,
		Value:         msg.Value,
		TimestampType: kafka.TimestampCreateTime,
		Opaque:        msg,
	}

	callback := func(m *Message) {
		assert.Equal(t, km.Key, m.Key)
		assert.Equal(t, km.Value, m.Value)
		assert.Equal(t, Success, m.Status)
	}
	producer, mockKafka := newTestProducer(t, DeliveryCallback(callback))

	msg.AddCallback(callback)

	mockKafka.On("Produce", km, mock.Anything).Run(func(args mock.Arguments) {
		go func() {
			args.Get(1).(chan kafka.Event) <- km
			cancel()
		}()
	}).Return(nil)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		err := producer.Run(ctx)
		assert.NoError(t, err)
	}()

	err := producer.Publish(context.Background(), msg)
	assert.NoError(t, err)

	wg.Wait()

	mockKafka.AssertExpectations(t)
}

func TestProducerPublishError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	msg := newFakeMessage()
	km := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &msg.Topic,
			Partition: kafka.PartitionAny,
		},
		Key:           msg.Key,
		Value:         msg.Value,
		TimestampType: kafka.TimestampCreateTime,
		Opaque:        msg,
	}
	expectErr := fmt.Errorf("kafka error")

	callback := func(m *Message) {
		assert.Equal(t, km.Key, m.Key)
		assert.Equal(t, km.Value, m.Value)
		assert.Equal(t, Fail, m.Status)
	}
	producer, mockKafka := newTestProducer(t, DeliveryCallback(callback))

	msg.AddCallback(callback)

	mockKafka.On("Produce", km, mock.Anything).Run(func(args mock.Arguments) {
		go func() {
			kmWithError := *km
			kmWithError.TopicPartition.Error = expectErr

			args.Get(1).(chan kafka.Event) <- &kmWithError

			cancel()
		}()
	}).Return(nil)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		err := producer.Run(ctx)
		assert.NoError(t, err)
	}()

	err := producer.Publish(context.Background(), msg)
	assert.Equal(t, expectErr, err)

	wg.Wait()

	mockKafka.AssertExpectations(t)
}

func TestProducerPublishRetryableError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	msg := newFakeMessage()
	km := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &msg.Topic,
			Partition: kafka.PartitionAny,
		},
		Key:           msg.Key,
		Value:         msg.Value,
		TimestampType: kafka.TimestampCreateTime,
		Opaque:        msg,
	}
	expectErr := fmt.Errorf("enqueue error")

	callback := func(m *Message) {
		assert.Equal(t, km.Key, m.Key)
		assert.Equal(t, km.Value, m.Value)
		assert.Equal(t, Fail, m.Status)
	}
	producer, mockKafka := newTestProducer(t, DeliveryCallback(callback))
	msg.AddCallback(callback)

	mockKafka.On("Produce", km, mock.Anything).Run(func(args mock.Arguments) {
		go func() {
			cancel()
		}()
	}).Return(expectErr)

	err := producer.Publish(context.Background(), msg)
	assert.ErrorIs(t, err, expectErr)
	assert.ErrorContains(t, err, ErrRetryable.Error())

	err = producer.Run(ctx)
	assert.NoError(t, err)

	mockKafka.AssertExpectations(t)
}

func TestProducerAsyncPublish(t *testing.T) {
	msg := newFakeMessage()
	expect := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &msg.Topic,
			Partition: kafka.PartitionAny,
		},
		Key:           msg.Key,
		Value:         msg.Value,
		TimestampType: kafka.TimestampCreateTime,
		Opaque:        msg,
	}
	ctx, cancel := context.WithCancel(context.Background())

	callback := func(m *Message) {
		assert.Equal(t, expect.Key, m.Key)
		assert.Equal(t, expect.Value, m.Value)
		assert.Equal(t, Success, m.Status)
	}
	producer, mockKafka := newTestProducer(t, DeliveryCallback(callback))

	msg.AddCallback(callback)

	produceCh := make(chan *kafka.Message, 1)
	mockKafka.On("ProduceChannel").Return(produceCh)

	err := producer.AsyncPublish(context.Background(), msg)
	assert.NoError(t, err)

	got := <-produceCh
	assert.EqualValues(t, expect, got)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		producer.events <- expect

		cancel()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		err := producer.Run(ctx)
		assert.NoError(t, err)
	}()

	wg.Wait()

	mockKafka.AssertExpectations(t)
}

func TestProducerAsyncPublishError(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	msg := newFakeMessage()
	expect := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &msg.Topic,
			Partition: kafka.PartitionAny,
		},
		Key:           msg.Key,
		Value:         msg.Value,
		TimestampType: kafka.TimestampCreateTime,
		Opaque:        msg,
	}

	callback := func(m *Message) {
		assert.Equal(t, expect.Key, m.Key)
		assert.Equal(t, expect.Value, m.Value)
		assert.Equal(t, Fail, m.Status)
	}
	producer, mockKafka := newTestProducer(t, DeliveryCallback(callback))

	msg.AddCallback(callback)

	produceCh := make(chan *kafka.Message, 1)
	mockKafka.On("ProduceChannel").Return(produceCh)

	err := producer.AsyncPublish(context.Background(), msg)
	assert.NoError(t, err)

	got := <-produceCh
	assert.EqualValues(t, expect, got)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		kmWithError := *expect
		kmWithError.TopicPartition.Error = fmt.Errorf("kafka error")

		producer.events <- &kmWithError

		cancel()
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		err := producer.Run(ctx)
		assert.NoError(t, err)
	}()

	wg.Wait()

	mockKafka.AssertExpectations(t)
}

func TestProducerMiddlewareExecutionOrder(t *testing.T) {
	producer, mockKafka := newTestProducer(t)

	ctx, cancel := context.WithCancel(context.Background())
	msg := newFakeMessage()
	km := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &msg.Topic,
			Partition: kafka.PartitionAny,
		},
		Key:           msg.Key,
		Value:         msg.Value,
		TimestampType: kafka.TimestampCreateTime,
		Opaque:        msg,
	}

	mockKafka.On("Produce", km, mock.Anything).Run(func(args mock.Arguments) {
		go func() {
			args.Get(1).(chan kafka.Event) <- km
			cancel()
		}()
	}).Return(nil)

	var preExec []string
	var postExec []string

	middlewares := []MiddlewareFunc{
		testMiddleware("middleware1", &preExec, &postExec),
		testMiddleware("middleware2", &preExec, &postExec),
		testMiddleware("middleware3", &preExec, &postExec),
	}

	producer.Use(middlewares...)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		err := producer.Run(ctx)
		assert.NoError(t, err)
	}()

	err := producer.Publish(context.Background(), msg)
	assert.NoError(t, err)

	wg.Wait()

	assert.Equal(t, []string{"middleware1", "middleware2", "middleware3"}, preExec)
	assert.Equal(t, []string{"middleware3", "middleware2", "middleware1"}, postExec)

	mockKafka.AssertExpectations(t)
}

func TestProducerIgnoreOpaqueMessage(t *testing.T) {
	producer, mockKafka := newTestProducer(t)

	ctx, cancel := context.WithCancel(context.Background())
	msg := newFakeMessage()
	km := &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &msg.Topic,
			Partition: kafka.PartitionAny,
		},
		Key:           msg.Key,
		Value:         msg.Value,
		TimestampType: kafka.TimestampCreateTime,
		Opaque:        msg,
	}

	mockKafka.On("Produce", km, mock.Anything).Run(func(args mock.Arguments) {
		go func() {
			e := km
			e.Opaque = nil

			args.Get(1).(chan kafka.Event) <- km
			cancel()
		}()
	}).Return(nil)

	called := false
	msg.AddCallback(func(m *Message) {
		called = true
	})

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		err := producer.Run(ctx)
		assert.NoError(t, err)
	}()

	err := producer.Publish(context.Background(), msg)
	assert.NoError(t, err)

	wg.Wait()

	assert.False(t, called)
	mockKafka.AssertExpectations(t)
}

func newTestProducer(t *testing.T, opts ...Option) (*Producer, *MockProducerClient) {
	mockKafka := &MockProducerClient{}

	mockKafka.On("Events").Return(make(chan kafka.Event, 1))
	mockKafka.On("Close").Return()
	mockKafka.On("Flush", mock.Anything).Return(0)

	opts = append(opts, testBrokers, mockProducerFunc(mockKafka), ShutdownTimeout(time.Second))

	producer, err := NewProducer(
		"producer-id",
		opts...,
	)
	require.NoError(t, err)
	require.NotNil(t, producer)

	return producer, mockKafka
}

func newFakeMessage() *Message {
	return &Message{
		Topic: testTopics[0],
		Key:   []byte("test-key"),
		Value: []byte("test-value"),
	}
}

func mockProducerFunc(mock *MockProducerClient) producerFunc {
	return func(configMap *kafka.ConfigMap) (producerClient, error) {
		return mock, nil
	}
}
