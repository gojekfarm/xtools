package xkafkaprom_test

import (
	"context"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/gojekfarm/xtools/xkafka"
	"github.com/gojekfarm/xtools/xprom/xkafkaprom"
)

var handler = xkafka.HandlerFunc(func(ctx context.Context, m *xkafka.Message) error {
	// Handle message.
	return nil
})

func ExampleCollector_ConsumerMiddleware() {
	consumer, _ := xkafka.NewConsumer(
		"test-group",
		handler,
		xkafka.Brokers{"localhost:9092"},
		xkafka.Topics{"test-topic"},
	)

	reg := prometheus.NewRegistry()
	collector := xkafkaprom.RegisterCollector(
		reg,
		xkafkaprom.LatencyBuckets{0.1, 0.5, 1, 2, 5},
		xkafkaprom.Address("localhost:9092"),
		xkafkaprom.ErrorClassifer(func(err error) string {
			// Classify errors.
			return "CustomError"
		}),
	)

	consumer.Use(collector.ConsumerMiddleware())

	// Start consuming messages.
}

func ExampleCollector_ConsumerMiddleware_multipleConsumers() {
	reg := prometheus.NewRegistry()
	collector := xkafkaprom.RegisterCollector(
		reg,
		xkafkaprom.LatencyBuckets{0.1, 0.5, 1, 2, 5},
		xkafkaprom.ErrorClassifer(func(err error) string {
			// Classify errors.
			return "CustomError"
		}),
	)

	consumer1, _ := xkafka.NewConsumer(
		"test-group-1",
		handler,
		xkafka.Brokers{"localhost:9092"},
		xkafka.Topics{"test-topic-1"},
	)

	consumer1.Use(collector.ConsumerMiddleware(
		xkafkaprom.Address("localhost:9092"),
	))

	consumer2, _ := xkafka.NewConsumer(
		"test-group-2",
		handler,
		xkafka.Brokers{"another-host:9092"},
		xkafka.Topics{"test-topic-2"},
	)

	consumer2.Use(collector.ConsumerMiddleware(
		xkafkaprom.Address("another-host:9092"),
	))

	// Start consuming messages.
}

func ExampleCollector_ProducerMiddleware() {
	producer, _ := xkafka.NewProducer(
		"test-publisher",
		xkafka.Brokers{"localhost:9092"},
	)

	reg := prometheus.NewRegistry()
	collector := xkafkaprom.RegisterCollector(
		reg,
		xkafkaprom.LatencyBuckets{0.1, 0.5, 1, 2, 5},
		xkafkaprom.Address("localhost:9092"),
		xkafkaprom.ErrorClassifer(func(err error) string {
			// Classify errors.
			return "CustomError"
		}),
	)

	producer.Use(collector.ProducerMiddleware())

	// Produce messages.
}

func ExampleCollector_ProducerMiddleware_multipleProducers() {
	reg := prometheus.NewRegistry()
	collector := xkafkaprom.RegisterCollector(
		reg,
		xkafkaprom.LatencyBuckets{0.1, 0.5, 1, 2, 5},
		xkafkaprom.ErrorClassifer(func(err error) string {
			// Classify errors.
			return "CustomError"
		}),
	)

	producer1, _ := xkafka.NewProducer(
		"test-publisher-1",
		xkafka.Brokers{"localhost:9092"},
	)

	producer1.Use(collector.ProducerMiddleware(
		xkafkaprom.Address("localhost:9092"),
	))

	producer2, _ := xkafka.NewProducer(
		"test-publisher-2",
		xkafka.Brokers{"another-host:9092"},
	)

	producer2.Use(collector.ProducerMiddleware(
		xkafkaprom.Address("another-host:9092"),
	))

	// Produce messages.
}
