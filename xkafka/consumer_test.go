package xkafka_test

import (
	"context"
	"testing"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/gojekfarm/xtools/xkafka"
)

func TestNewConsumer(t *testing.T) {
	consumer, err := xkafka.NewConsumer(
		"consumer-id",
		xkafka.Topics([]string{"test"}),
		xkafka.Brokers([]string{"localhost:9092"}),
		xkafka.ConfigMap(xkafka.ConfigMap{
			"enable.auto.commit": false,
		}),
	)
	assert.NoError(t, err)
	assert.NotNil(t, consumer)

	noopMiddleware := xkafka.MiddlewareFunc(func(next xkafka.Handler) xkafka.Handler {
		return next
	})

	consumer.Use(noopMiddleware)
}

func TestGetMetadata(t *testing.T) {
	mock := &MockKafkaConsumer{}
	consumer, _ := xkafka.NewConsumer(
		"consumer-id",
		xkafka.ConsumerFunc(mockConsumerFunc(mock)),
		xkafka.WithMetadataRequestTimeout(10*time.Second),
	)
	require.NotNil(t, consumer)

	mock.On("GetMetadata", (*string)(nil), false, 10000).Return(&kafka.Metadata{}, nil)

	metadata, err := consumer.GetMetadata()
	assert.NoError(t, err)
	assert.NotNil(t, metadata)

	mock.AssertExpectations(t)
}

func TestConsumer(t *testing.T) {
	kafka := &MockKafkaConsumer{}
	consumer, _ := xkafka.NewConsumer(
		"consumer-id",
		xkafka.Topics([]string{topic}),
		xkafka.Brokers([]string{"localhost:9092"}),
		xkafka.ConsumerFunc(mockConsumerFunc(kafka)),
		xkafka.WithMetadataRequestTimeout(10*time.Second),
		xkafka.WithPollTimeout(1*time.Second),
	)
	require.NotNil(t, consumer)

	km := newKafkaMessage()
	ctx, cancel := context.WithCancel(context.Background())
	handler := xkafka.HandlerFunc(func(ctx context.Context, msg *xkafka.Message) error {
		assert.Equal(t, topic, msg.Topic)
		assert.Equal(t, "key", string(msg.Key))
		assert.Equal(t, "value", string(msg.Value))

		cancel()
		return nil
	})

	kafka.On("SubscribeTopics", []string{topic}, mock.Anything).Return(nil)
	kafka.On("ReadMessage", 1*time.Second).Return(km, nil)

	consumer.Start(ctx, handler)

	kafka.AssertExpectations(t)
}

func mockConsumerFunc(mock *MockKafkaConsumer) xkafka.ConsumerFunc {
	return func(configMap *kafka.ConfigMap) (xkafka.KafkaConsumer, error) {
		return mock, nil
	}
}

func newKafkaMessage() *kafka.Message {
	return &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &topic,
			Partition: 1,
		},
		Value:     []byte("value"),
		Key:       []byte("key"),
		Timestamp: time.Date(2020, 1, 1, 23, 59, 59, 0, time.UTC),
	}
}
