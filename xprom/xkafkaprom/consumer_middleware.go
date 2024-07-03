package prometheus

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/gojekfarm/xtools/xkafka"
	"github.com/gojekfarm/xtools/xprom/semconv"
)

var consumerCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
	Name: semconv.MessagingClientOperationDuration,
	Help: "Total number of messages processed by consumer, partitioned by group, topic, partition and status.",
	ConstLabels: prometheus.Labels{
		semconv.MessagingSystem:        semconv.SystemKafka,
		semconv.MessagingOperationName: semconv.OperationConsume,
	},
}, []string{
	semconv.MessagingKafkaConsumerGroup,
	semconv.MessagingKafkaTopic,
	semconv.MessagingKafkaPartition,
	semconv.MessagingKafkaMessageStatus,
})

var consumerInflight = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: semconv.MessagingInflightMessages,
	Help: "Messages currently being processed by consumer. Partitioned by group, topic and partition.",
	ConstLabels: prometheus.Labels{
		semconv.MessagingSystem:        semconv.SystemKafka,
		semconv.MessagingOperationName: semconv.OperationConsume,
	},
}, []string{
	semconv.MessagingKafkaConsumerGroup,
	semconv.MessagingKafkaTopic,
	semconv.MessagingKafkaPartition,
})

var consumerDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name: semconv.MessagingClientOperationDuration,
	Help: "Message processing duration, partitioned by group, topic, partition and status.",
	ConstLabels: prometheus.Labels{
		semconv.MessagingSystem: semconv.SystemKafka,
	},
}, []string{
	semconv.MessagingKafkaConsumerGroup,
	semconv.MessagingKafkaTopic,
	semconv.MessagingKafkaPartition,
	semconv.MessagingKafkaMessageStatus,
})

var allConsumerCollectors = []prometheus.Collector{
	consumerCounter,
	consumerInflight,
	consumerDuration,
}

// ConsumerMiddleware adds prometheus instrumentation for xkafka.Consumer.
func ConsumerMiddleware(next xkafka.Handler) xkafka.Handler {
	return xkafka.HandlerFunc(func(ctx context.Context, m *xkafka.Message) error {
		startTime := time.Now()
		partition := fmt.Sprintf("%d", m.Partition)
		labels := []string{m.Group, m.Topic, partition}

		inflight := consumerInflight.WithLabelValues(labels...)
		inflight.Inc()

		m.AddCallback(func(ackMsg *xkafka.Message) {
			labels := []string{
				ackMsg.Group, ackMsg.Topic, partition, ackMsg.Status.String(),
			}

			consumerCounter.WithLabelValues(labels...).
				Inc()

			consumerDuration.WithLabelValues(labels...).
				Observe(time.Since(startTime).Seconds())

			inflight.Dec()
		})

		return next.Handle(ctx, m)
	})
}
