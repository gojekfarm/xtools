package xkafka

import (
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

const (
	atleastLeaderAck = 1
	partitioner      = "consistent_random"
)

// Brokers is an option to set the brokers.
type Brokers []string

func (b Brokers) apply(o *options) { o.brokers = b }

// Topics is an option to set the topics.
type Topics []string

func (t Topics) apply(o *options) { o.topics = t }

// ConfigMap is an option to set the config map.
type ConfigMap map[string]interface{}

func (cm ConfigMap) apply(o *options) {
	for k, v := range cm {
		_ = o.configMap.SetKey(k, v)
	}
}

type options struct {
	// common options
	brokers         []string
	configMap       kafka.ConfigMap
	errorHandler    ErrorHandler
	shutdownTimeout time.Duration

	// consumer options
	consumerFn             ConsumerFunc
	topics                 []string
	metadataRequestTimeout time.Duration
	pollTimeout            time.Duration
	concurrency            int

	// producer options
	producerFn ProducerFunc
}

func defaultConsumerOptions() options {
	return options{
		consumerFn:             DefaultConsumerFunc,
		topics:                 []string{},
		brokers:                []string{},
		configMap:              kafka.ConfigMap{},
		errorHandler:           NoopErrorHandler,
		metadataRequestTimeout: 10 * time.Second,
		pollTimeout:            10 * time.Second,
		shutdownTimeout:        1 * time.Second,
		concurrency:            1,
	}
}

func defaultProducerOptions() options {
	return options{
		producerFn: DefaultProducerFunc,
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

// Option is an interface for consumer options.
type Option interface{ apply(*options) }

type optionFunc func(*options)

func (f optionFunc) apply(o *options) { f(o) }

// WithMetadataRequestTimeout sets the metadata request timeout.
func WithMetadataRequestTimeout(timeout time.Duration) Option {
	return optionFunc(func(o *options) {
		o.metadataRequestTimeout = timeout
	})
}

// WithPollTimeout sets the poll timeout.
func WithPollTimeout(timeout time.Duration) Option {
	return optionFunc(func(o *options) {
		o.pollTimeout = timeout
	})
}

// WithShutdownTimeout sets the shutdown timeout.
func WithShutdownTimeout(timeout time.Duration) Option {
	return optionFunc(func(o *options) {
		o.shutdownTimeout = timeout
	})
}

// WithConcurrency sets the concurrency level.
// Used with NewConcurrent only.
func WithConcurrency(concurrency int) Option {
	return optionFunc(func(o *options) {
		o.concurrency = concurrency
	})
}
