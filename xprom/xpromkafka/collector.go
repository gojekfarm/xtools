package xpromkafka

import (
	"context"
	"fmt"
	"time"

	"github.com/prometheus/client_golang/prometheus"

	"github.com/gojekfarm/xtools/xkafka"
	"github.com/gojekfarm/xtools/xprom/semconv"
)

var (
	defaultLatencyBuckets = []float64{
		0.001, // 1ms
		0.002, // 2ms
		0.005, // 5ms
		0.01,  // 10ms
		0.02,  // 20ms
		0.05,  // 50ms
		0.1,   // 100ms
		0.2,   // 200ms
		0.5,   // 500ms
		1,     // 1s
		2,     // 2s
		5,     // 5s
		10,    // 10s
	}
)

// Collector provides metrics for xkafka.Producer and xkafka.Consumer.
type Collector struct {
	opts      options
	duration  *prometheus.HistogramVec
	inflight  *prometheus.GaugeVec
	published *prometheus.CounterVec
	consumed  *prometheus.CounterVec
}

// NewCollector creates a new Collector.
func NewCollector(opts ...Option) *Collector {
	o := options{
		latencyBuckets: defaultLatencyBuckets,
	}

	for _, opt := range opts {
		opt.apply(&o)
	}

	constLabels := prometheus.Labels{
		semconv.MessagingSystem: semconv.SystemKafka,
	}

	return &Collector{
		opts: o,
		duration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:        semconv.MessagingClientOperationDuration,
			Help:        "Message processing duration.",
			ConstLabels: constLabels,
			Buckets:     o.latencyBuckets,
		}, []string{
			semconv.MessagingOperationName,
			semconv.ServerAddress,
			semconv.ServerPort,
			semconv.MessagingKafkaConsumerGroup,
			semconv.MessagingKafkaTopic,
			semconv.MessagingKafkaPartition,
			semconv.MessagingKafkaMessageStatus,
			semconv.ErrorType,
		}),
		inflight: prometheus.NewGaugeVec(prometheus.GaugeOpts{
			Name:        semconv.MessagingInflightMessages,
			Help:        "Messages currently being processed.",
			ConstLabels: constLabels,
		}, []string{
			semconv.MessagingOperationName,
			semconv.ServerAddress,
			semconv.ServerPort,
			semconv.MessagingKafkaConsumerGroup,
			semconv.MessagingKafkaTopic,
			semconv.MessagingKafkaPartition,
		}),
		published: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name:        semconv.MessagingClientPublishedMessages,
			Help:        "Messages published.",
			ConstLabels: constLabels,
		}, []string{
			semconv.MessagingOperationName,
			semconv.ServerAddress,
			semconv.ServerPort,
			semconv.MessagingKafkaConsumerGroup,
			semconv.MessagingKafkaPartition,
			semconv.MessagingKafkaTopic,
			semconv.MessagingKafkaMessageStatus,
			semconv.ErrorType,
		}),
		consumed: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name:        semconv.MessagingClientConsumedMessages,
			Help:        "Messages consumed.",
			ConstLabels: constLabels,
		}, []string{
			semconv.MessagingOperationName,
			semconv.ServerAddress,
			semconv.ServerPort,
			semconv.MessagingKafkaConsumerGroup,
			semconv.MessagingKafkaTopic,
			semconv.MessagingKafkaPartition,
			semconv.MessagingKafkaMessageStatus,
			semconv.ErrorType,
		}),
	}
}

// Register registers the metrics with the provided registry.
func (c *Collector) Register(registry prometheus.Registerer) error {
	if err := registry.Register(c.duration); err != nil {
		return err
	}

	if err := registry.Register(c.inflight); err != nil {
		return err
	}

	if err := registry.Register(c.published); err != nil {
		return err
	}

	if err := registry.Register(c.consumed); err != nil {
		return err
	}

	return nil
}

// ConsumerMiddleware returns a middleware that instruments xkafka.Consumer.
// Options passed to this function will override the Collector options.
func (c *Collector) ConsumerMiddleware(opts ...Option) xkafka.MiddlewareFunc {
	mwopts := &options{
		errFn:   c.opts.errFn,
		address: c.opts.address,
		port:    c.opts.port,
	}

	for _, opt := range opts {
		opt.apply(mwopts)
	}

	return func(next xkafka.Handler) xkafka.Handler {
		return xkafka.HandlerFunc(func(ctx context.Context, msg *xkafka.Message) error {
			start := time.Now()
			labels := prometheus.Labels{
				semconv.MessagingOperationName:      semconv.OperationConsume,
				semconv.ServerAddress:               mwopts.address,
				semconv.MessagingKafkaConsumerGroup: msg.Group,
				semconv.MessagingKafkaTopic:         msg.Topic,
				semconv.MessagingKafkaPartition:     fmt.Sprintf("%d", msg.Partition),
			}

			if mwopts.port != 0 {
				labels[semconv.ServerPort] = fmt.Sprintf("%d", mwopts.port)
			}

			inflight := c.inflight.With(labels)

			inflight.Inc()
			defer inflight.Dec()

			msg.AddCallback(func(ackMsg *xkafka.Message) {
				labels := labels

				labels[semconv.MessagingKafkaMessageStatus] = ackMsg.Status.String()
				labels[semconv.ErrorType] = ""

				if ackMsg.Err() != nil && mwopts.errFn != nil {
					labels[semconv.ErrorType] = mwopts.errFn(ackMsg.Err())
				}

				c.duration.With(labels).Observe(time.Since(start).Seconds())
				c.consumed.With(labels).Inc()
			})

			return next.Handle(ctx, msg)
		})
	}
}

// ProducerMiddleware returns a middleware that instruments xkafka.Producer.
// Options passed to this function will override the Collector options.
func (c *Collector) ProducerMiddleware(opts ...Option) xkafka.MiddlewareFunc {
	mwopts := &options{
		errFn:   c.opts.errFn,
		address: c.opts.address,
		port:    c.opts.port,
	}

	for _, opt := range opts {
		opt.apply(mwopts)
	}

	return func(next xkafka.Handler) xkafka.Handler {
		return xkafka.HandlerFunc(func(ctx context.Context, msg *xkafka.Message) error {
			start := time.Now()
			labels := prometheus.Labels{
				semconv.MessagingOperationName:      semconv.OperationPublish,
				semconv.ServerAddress:               mwopts.address,
				semconv.MessagingKafkaTopic:         msg.Topic,
				semconv.MessagingKafkaPartition:     "",
				semconv.MessagingKafkaConsumerGroup: msg.Group,
			}

			if mwopts.port != 0 {
				labels[semconv.ServerPort] = fmt.Sprintf("%d", mwopts.port)
			}

			inflight := c.inflight.With(labels)

			inflight.Inc()
			defer inflight.Dec()

			msg.AddCallback(func(ackMsg *xkafka.Message) {
				labels := labels

				if ackMsg.Partition != -1 {
					labels[semconv.MessagingKafkaPartition] = fmt.Sprintf("%d", ackMsg.Partition)
				}

				labels[semconv.MessagingKafkaMessageStatus] = ackMsg.Status.String()
				labels[semconv.ErrorType] = ""

				if ackMsg.Err() != nil && mwopts.errFn != nil {
					labels[semconv.ErrorType] = mwopts.errFn(ackMsg.Err())
				}

				c.duration.With(labels).Observe(time.Since(start).Seconds())
				c.published.With(labels).Inc()
			})

			return next.Handle(ctx, msg)
		})
	}
}
