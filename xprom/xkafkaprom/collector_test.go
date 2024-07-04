package xkafkaprom

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"

	"github.com/gojekfarm/xtools/xkafka"
)

func TestConsumerMiddleware(t *testing.T) {
	msg := &xkafka.Message{
		Topic:     "test-topic",
		Group:     "test-group",
		Key:       []byte("key"),
		Value:     []byte("value"),
		Partition: 12,
	}

	handler := xkafka.HandlerFunc(func(ctx context.Context, m *xkafka.Message) error {
		err := errors.New("some error")

		m.AckFail(err)

		return err
	})

	reg := prometheus.NewRegistry()
	collector := RegisterCollector(
		reg,
		LatencyBuckets{0.1, 0.5, 1, 2, 5},
	)

	instrumentedHandler := collector.ConsumerMiddleware(
		Address("localhost"),
		Port(9092),
		ErrorClassifer(func(err error) string {
			return "CustomError"
		}),
	).Middleware(handler)

	err := instrumentedHandler.Handle(context.TODO(), msg)
	assert.Error(t, err)

	expectedMetrics := []string{
		"messaging_inflight_messages",
		"messaging_client_consumed_messages",
	}
	expected := `
	# HELP messaging_client_consumed_messages Messages consumed.
	# TYPE messaging_client_consumed_messages counter
	messaging_client_consumed_messages{error_type="CustomError",messaging_consumer_group_name="test-group",messaging_destination_name="test-topic",messaging_destination_partition_id="12",messaging_kafka_message_status="FAIL",messaging_operation_name="consume",messaging_system="kafka",server_address="localhost",server_port="9092"} 1
	# HELP messaging_inflight_messages Messages currently being processed.
	# TYPE messaging_inflight_messages gauge
	messaging_inflight_messages{messaging_consumer_group_name="test-group",messaging_destination_name="test-topic",messaging_destination_partition_id="12",messaging_operation_name="consume",messaging_system="kafka",server_address="localhost",server_port="9092"} 0
	`

	err = testutil.GatherAndCompare(reg, strings.NewReader(expected), expectedMetrics...)
	assert.NoError(t, err)
}

func TestProducerMiddleware(t *testing.T) {
	msg := &xkafka.Message{
		Topic: "test-topic",
		Group: "test-group",
		Key:   []byte("key"),
		Value: []byte("value"),
	}

	handler := xkafka.HandlerFunc(func(ctx context.Context, m *xkafka.Message) error {
		m.Partition = 12
		err := errors.New("some error")
		m.AckFail(err)
		return err
	})

	reg := prometheus.NewRegistry()
	collector := RegisterCollector(
		reg,
		LatencyBuckets{0.1, 0.5, 1, 2, 5},
	)

	instrumentedHandler := collector.ProducerMiddleware(
		Address("localhost"),
		Port(9092),
		ErrorClassifer(func(err error) string {
			return "CustomError"
		}),
	).Middleware(handler)

	err := instrumentedHandler.Handle(context.TODO(), msg)
	assert.Error(t, err)

	expectedMetrics := []string{
		"messaging_inflight_messages",
		"messaging_client_published_messages",
	}
	expected := `# HELP messaging_client_published_messages Messages published.
	# TYPE messaging_client_published_messages counter
	messaging_client_published_messages{error_type="CustomError",messaging_consumer_group_name="test-group",messaging_destination_name="test-topic",messaging_destination_partition_id="12",messaging_kafka_message_status="FAIL",messaging_operation_name="publish",messaging_system="kafka",server_address="localhost",server_port="9092"} 1
	# HELP messaging_inflight_messages Messages currently being processed.
	# TYPE messaging_inflight_messages gauge
	messaging_inflight_messages{messaging_consumer_group_name="test-group",messaging_destination_name="test-topic",messaging_destination_partition_id="",messaging_operation_name="publish",messaging_system="kafka",server_address="localhost",server_port="9092"} 0
	`

	err = testutil.GatherAndCompare(reg, strings.NewReader(expected), expectedMetrics...)
	assert.NoError(t, err)
}
