package xkafka

import (
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

const (
	atleastLeaderAck = 1
	partitioner      = "consistent_random"
)

// Option is an interface for consumer options.
type Option interface{ apply(*options) }

// Brokers is an option to set the brokers.
type Brokers []string

func (b Brokers) apply(o *options) { o.brokers = b }

// Topics is an option to set the topics.
type Topics []string

func (t Topics) apply(o *options) { o.topics = t }

// ConfigMap is an option to set kafka.ConfigMap.
type ConfigMap map[string]interface{}

func (cm ConfigMap) apply(o *options) {
	for k, v := range cm {
		_ = o.configMap.SetKey(k, v)
	}
}

// ShutdownTimeout defines the timeout for the consumer/producer to shutdown.
type ShutdownTimeout time.Duration

func (st ShutdownTimeout) apply(o *options) { o.shutdownTimeout = time.Duration(st) }

// PollTimeout defines the timeout for the consumer read timeout.
type PollTimeout time.Duration

func (pt PollTimeout) apply(o *options) { o.pollTimeout = time.Duration(pt) }

// MetadataTimeout defines the timeout for the consumer metadata request.
type MetadataTimeout time.Duration

func (mt MetadataTimeout) apply(o *options) { o.metadataTimeout = time.Duration(mt) }

// Concurrency defines the concurrency of the consumer.
type Concurrency int

func (c Concurrency) apply(o *options) { o.concurrency = int(c) }

// DeliveryCallback is a callback function triggered for every published message.
// Works only for xkafka.Producer.
type DeliveryCallback AckFunc

func (d DeliveryCallback) apply(o *options) { o.deliveryCb = d }

// ManualCommit disables the auto commit and calls the `Commit` after every
// message is marked as `Success` or `Skip` by the handler.
//
// Works only for xkafka.Consumer.
//
// WARNING: Using this option will increase the message processing time,
// because of the synchronous `Commit` for every message.
type ManualCommit bool

func (mc ManualCommit) apply(o *options) { o.manualCommit = bool(mc) }

type options struct {
	// common options
	brokers         []string
	configMap       kafka.ConfigMap
	errorHandler    ErrorHandler
	shutdownTimeout time.Duration

	// consumer options
	consumerFn      consumerFunc
	topics          []string
	metadataTimeout time.Duration
	pollTimeout     time.Duration
	concurrency     int
	manualCommit    bool

	// producer options
	producerFn producerFunc
	deliveryCb DeliveryCallback
}

func defaultConsumerOptions() options {
	return options{
		consumerFn:      defaultConsumerFunc,
		topics:          []string{},
		brokers:         []string{},
		configMap:       kafka.ConfigMap{},
		errorHandler:    NoopErrorHandler,
		metadataTimeout: 10 * time.Second,
		pollTimeout:     10 * time.Second,
		shutdownTimeout: 1 * time.Second,
		concurrency:     1,
	}
}

func defaultProducerOptions() options {
	return options{
		producerFn: defaultProducerFunc,
		brokers:    []string{},
		configMap: kafka.ConfigMap{
			"default.topic.config": kafka.ConfigMap{
				"acks":        atleastLeaderAck,
				"partitioner": partitioner,
			},
		},
		errorHandler: NoopErrorHandler,
	}
}
