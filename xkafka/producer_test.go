package xkafka

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/stretchr/testify/assert"
	mock "github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestProducerPublish(t *testing.T) {
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

	callback := func(m *Message) {
		assert.Equal(t, km.Key, m.Key)
		assert.Equal(t, km.Value, m.Value)
		assert.Equal(t, Success, m.Status)
	}
	msg.AddCallback(callback)

	mockKafka.On("Produce", km, mock.Anything).Run(func(args mock.Arguments) {
		go func() {
			args.Get(1).(chan kafka.Event) <- km
			cancel()
		}()
	}).Return(nil)

	err := producer.Publish(context.Background(), msg)
	assert.NoError(t, err)

	err = producer.Run(ctx)
	assert.NoError(t, err)

	mockKafka.AssertExpectations(t)
}

func TestProducerPublishError(t *testing.T) {
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
	expectErr := fmt.Errorf("kafka error")

	msg.AddCallback(func(m *Message) {
		assert.Equal(t, km.Key, m.Key)
		assert.Equal(t, km.Value, m.Value)
		assert.Equal(t, Fail, m.Status)
	})

	mockKafka.On("Produce", km, mock.Anything).Run(func(args mock.Arguments) {
		go func() {
			kmWithError := *km
			kmWithError.TopicPartition.Error = expectErr

			args.Get(1).(chan kafka.Event) <- &kmWithError

			cancel()
		}()
	}).Return(nil)

	err := producer.Publish(context.Background(), msg)
	assert.Equal(t, expectErr, err)

	err = producer.Run(ctx)
	assert.NoError(t, err)

	mockKafka.AssertExpectations(t)
}

func TestProducerPublishRetryableError(t *testing.T) {
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
	expectErr := fmt.Errorf("enqueue error")

	callback := func(m *Message) {
		assert.Equal(t, km.Key, m.Key)
		assert.Equal(t, km.Value, m.Value)
		assert.Equal(t, Fail, m.Status)
	}
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
	producer, mockKafka := newTestProducer(t)

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

	producer.config.deliveryCb = DeliveryCallback(callback)

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
	producer, mockKafka := newTestProducer(t)

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

	producer.config.deliveryCb = DeliveryCallback(callback)

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

func newTestProducer(t *testing.T) (*Producer, *MockProducerClient) {
	mockKafka := &MockProducerClient{}

	mockKafka.On("Events").Return(make(chan kafka.Event, 1))
	mockKafka.On("Close").Return()
	mockKafka.On("Flush", mock.Anything).Return(0)

	producer, err := NewProducer(
		"producer-id",
		testBrokers,
		mockProducerFunc(mockKafka),
		ShutdownTimeout(1*time.Second),
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
