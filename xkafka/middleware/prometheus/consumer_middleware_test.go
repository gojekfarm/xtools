package prometheus

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/gojekfarm/xtools/xkafka"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestConsumerMiddleware(t *testing.T) {
	msg := &xkafka.Message{
		Topic:     "test-topic",
		Group:     "test-group",
		Partition: 2,
		Key:       []byte("key"),
		Value:     []byte("value"),
	}

	consumerHandler := xkafka.HandlerFunc(func(ctx context.Context, m *xkafka.Message) error {
		time.Sleep(1 * time.Second)

		m.AckSuccess()

		return nil
	})

	reg := prometheus.NewRegistry()
	err := RegisterConsumerMetrics(reg)
	assert.NoError(t, err)

	instrumentedHandler := ConsumerMiddleware(consumerHandler)
	err = instrumentedHandler.Handle(context.TODO(), msg)
	assert.NoError(t, err)

	time.Sleep(50 * time.Millisecond)

	expected := `
	# HELP kafka_consumer_messages_in_flight Current number of messages being processed by consumer, partitioned by group.
	# TYPE kafka_consumer_messages_in_flight gauge
	kafka_consumer_messages_in_flight{group="test-group",topic="test-topic"} 0
	# HELP kafka_consumer_messages_total How many messages consumed, partitioned by group and status.
	# TYPE kafka_consumer_messages_total counter
	kafka_consumer_messages_total{group="test-group",status="SUCCESS",topic="test-topic"} 1
`
	expectedMetrics := []string{
		"kafka_consumer_messages_in_flight",
		"kafka_consumer_messages_total",
	}

	err = testutil.GatherAndCompare(reg, strings.NewReader(expected), expectedMetrics...)
	assert.NoError(t, err)
}
