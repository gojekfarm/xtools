package prometheus

import (
	"context"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/gojekfarm/xtools/xkafka"
)

var consumerLabels = []string{
	LabelGroup,
	LabelStatus,
	LabelTopic,
}

var consumerLag = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name:      MetricMessageLagSeconds,
	Subsystem: SubsystemConsumer,
	Help:      "What is the lag of the consumer, partitioned by group, topic and partition.",
}, []string{LabelGroup, LabelTopic, LabelPartition})

var consumerCounter = prometheus.NewCounterVec(prometheus.CounterOpts{
	Name:      MetricMessagesTotal,
	Subsystem: SubsystemConsumer,
	Help:      "How many messages consumed, partitioned by group and status.",
}, consumerLabels)

var consumerInflightMessages = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Name:      MetricMessagesInFlight,
	Subsystem: SubsystemConsumer,
	Help:      "Current number of messages being processed by consumer, partitioned by group.",
}, []string{LabelGroup, LabelTopic})

var consumerProcessingDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Name:      MetricMessageDurationSeconds,
	Subsystem: SubsystemConsumer,
	Help:      "How long it took to process the message, partitioned by group and status.",
}, consumerLabels)

var allConsumerCollectors = []prometheus.Collector{
	consumerLag,
	consumerCounter,
	consumerInflightMessages,
	consumerProcessingDuration,
}

// ConsumerMiddleware adds prometheus instrumentation for xkafka.Consumer.
func ConsumerMiddleware(next xkafka.Handler) xkafka.Handler {
	return xkafka.HandlerFunc(func(ctx context.Context, m *xkafka.Message) error {
		startTime := time.Now()

		defer consumerLag.WithLabelValues(m.Group, m.Topic, string(m.Partition)).
			Observe(time.Since(m.Timestamp).Seconds())

		inflight := consumerInflightMessages.WithLabelValues(m.Group, m.Topic)
		inflight.Inc()

		m.AddCallback(func(ackMsg *xkafka.Message) {
			consumerCounter.
				WithLabelValues(ackMsg.Group, ackMsg.Status.String(), ackMsg.Topic).
				Inc()

			consumerProcessingDuration.
				WithLabelValues(ackMsg.Group, ackMsg.Status.String(), ackMsg.Topic).
				Observe(time.Since(startTime).Seconds())

			inflight.Dec()
		})

		return next.Handle(ctx, m)
	})
}
