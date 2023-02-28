package xkafka_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gojekfarm/xtools/xkafka"
)

func TestNewConsumer(t *testing.T) {
	consumer, err := xkafka.NewConsumer(
		"consumer-id",
		xkafka.WithConcurrency(10),
		xkafka.WithTopics("test"),
		xkafka.WithBrokers("localhost:9092"),
		xkafka.WithKafkaConfig("group.id", "test"),
	)
	assert.NoError(t, err)
	assert.NotNil(t, consumer)

	noopMiddleware := xkafka.MiddlewareFunc(func(next xkafka.Handler) xkafka.Handler {
		return next
	})

	consumer.Use(noopMiddleware)
}
