package xkafka_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

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
