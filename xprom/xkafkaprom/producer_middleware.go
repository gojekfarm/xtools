package prometheus

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/gojekfarm/xtools/xkafka"
	"github.com/gojekfarm/xtools/xprom/semconv"
)

var producerLabels = []string{
	semconv.MessagingKafkaTopic,
	semconv.MessagingKafkaMessageStatus,
}

var producerDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name: semconv.MessagingClientOperationDuration,
	ConstLabels: prometheus.Labels{
		semconv.MessagingSystem:        semconv.SystemKafka,
		semconv.MessagingOperationName: semconv.OperationPublish,
	},
	Help: "Message publishing duration, partitioned by topic and status.",
}, producerLabels)

var producerCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
	Name: semconv.MessagingClientPublishedMessages,
	ConstLabels: prometheus.Labels{
		semconv.MessagingSystem: semconv.SystemKafka,
	},
	Help: "Messages published, partitioned by topic and status.",
}, producerLabels)

var producerInflight = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name: semconv.MessagingInflightMessages,
	ConstLabels: prometheus.Labels{
		semconv.MessagingSystem:        semconv.SystemKafka,
		semconv.MessagingOperationName: semconv.OperationPublish,
	},
	Help: "Messages currently being published",
}, []string{
	semconv.MessagingKafkaTopic,
})

var allProducerCollectors = []prometheus.Collector{
	producerCounter,
	producerDuration,
	producerInflight,
}

// ProducerMiddleware adds prometheus instrumentation for xkafka.Producer.
func ProducerMiddleware(next xkafka.Handler) xkafka.Handler {
	return xkafka.HandlerFunc(func(ctx context.Context, m *xkafka.Message) error {
		start := time.Now()

		inflight := producerInflight.WithLabelValues(m.Topic)
		inflight.Inc()

		m.AddCallback(func(ackMsg *xkafka.Message) {
			producerCounter.
				WithLabelValues(ackMsg.Topic, ackMsg.Status.String()).
				Inc()

			producerDuration.
				WithLabelValues(ackMsg.Topic, ackMsg.Status.String()).
				Observe(time.Since(start).Seconds())

			inflight.Dec()
		})

		return next.Handle(ctx, m)
	})
}
