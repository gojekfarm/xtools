package prometheus

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"

	"github.com/gojekfarm/xtools/xkafka"
)

func TestProducerMiddleware(t *testing.T) {
	t.Parallel()

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

	expected := ``
	expectedMetrics := []string{
		"messaging_inflight_messages",
		"messaging_client_published_messages",
	}

	err = testutil.GatherAndCompare(reg, strings.NewReader(expected), expectedMetrics...)
	assert.NoError(t, err)
}
