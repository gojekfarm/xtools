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

func TestConsumerMiddleware(t *testing.T) {
	t.Parallel()

	msg := &xkafka.Message{
		Topic:     "test-topic",
		Group:     "test-group",
		Partition: 12,
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

	expectedMetrics := []string{
		"messaging_inflight_messages",
		"messaging_client_consumed_messages",
	}

	expected := ``

	err = testutil.GatherAndCompare(reg, strings.NewReader(expected), expectedMetrics...)
	assert.NoError(t, err)
}
