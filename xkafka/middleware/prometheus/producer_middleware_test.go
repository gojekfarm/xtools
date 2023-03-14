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

func TestProducerMiddleware(t *testing.T) {
	msg := &xkafka.Message{
		Topic: "test-topic",
		Group: "test-group",
		Key:   []byte("key"),
		Value: []byte("value"),
	}

	producerHandler := xkafka.HandlerFunc(func(ctx context.Context, m *xkafka.Message) error {
		m.AckSuccess()

		return nil
	})

	reg := prometheus.NewRegistry()
	err := RegisterProducerMetrics(reg)
	assert.NoError(t, err)

	instrumentedHandler := ProducerMiddleware(producerHandler)
	err = instrumentedHandler.Handle(context.TODO(), msg)
	assert.NoError(t, err)

	time.Sleep(50 * time.Millisecond)

	expected := `
	# HELP kafka_producer_messages_in_flight Current number of messages being processed by producer, partitioned by group and topic.
	# TYPE kafka_producer_messages_in_flight gauge
	kafka_producer_messages_in_flight{group="test-group",topic="test-topic"} 0
	# HELP kafka_producer_messages_total How many messages produced, partitioned by group, status and topic.
	# TYPE kafka_producer_messages_total counter
	kafka_producer_messages_total{group="test-group",status="SUCCESS",topic="test-topic"} 1
`
	expectedMetrics := []string{
		"kafka_producer_messages_in_flight",
		"kafka_producer_messages_total",
	}

	err = testutil.GatherAndCompare(reg, strings.NewReader(expected), expectedMetrics...)
	assert.NoError(t, err)
}
