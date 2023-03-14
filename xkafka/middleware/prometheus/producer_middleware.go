package prometheus

import (
	"context"
	"time"

	"github.com/gojekfarm/xtools/xkafka"
	"github.com/prometheus/client_golang/prometheus"
)

var producerLabels = []string{
	LabelGroup,
	LabelStatus,
	LabelTopic,
}

var producerCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
	Name:      MetricMessagesTotal,
	Subsystem: SubsystemProducer,
	Help:      "How many messages produced, partitioned by group, status and topic.",
}, producerLabels)

var producerRunDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name:      MetricMessageDurationSeconds,
	Subsystem: SubsystemProducer,
	Help:      "How long it took to process the message, partitioned by group, status and topic.",
}, producerLabels)

var producerInflightMessages = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name:      MetricMessagesInFlight,
	Subsystem: SubsystemProducer,
	Help:      "Current number of messages being processed by producer, partitioned by group and topic.",
}, []string{LabelGroup, LabelTopic})

var allProducerCollectors = []prometheus.Collector{
	producerCounter,
	producerRunDuration,
	producerInflightMessages,
}

// ProducerMiddleware adds prometheus instrumentation for xkafka.Producer.
func ProducerMiddleware(next xkafka.Handler) xkafka.Handler {
	return xkafka.HandlerFunc(func(ctx context.Context, m *xkafka.Message) error {
		start := time.Now()

		inflight := producerInflightMessages.WithLabelValues(m.Group, m.Topic)
		inflight.Inc()

		m.AddCallback(func(ackMsg *xkafka.Message) {
			producerCounter.
				WithLabelValues(ackMsg.Group, ackMsg.Status.String(), ackMsg.Topic).
				Inc()

			producerRunDuration.
				WithLabelValues(ackMsg.Group, ackMsg.Status.String(), ackMsg.Topic).
				Observe(time.Since(start).Seconds())

			inflight.Dec()
		})

		return next.Handle(ctx, m)
	})
}
