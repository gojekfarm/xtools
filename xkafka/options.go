package xkafka

import (
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

const (
	atleastLeaderAck = 1
	partitioner      = "consistent_random"
)

type options struct {
	// common options
	brokers         []string
	configMap       kafka.ConfigMap
	errorHandler    ErrorHandler
	shutdownTimeout time.Duration

	// consumer options
	topics                 []string
	metadataRequestTimeout time.Duration
	pollTimeout            time.Duration

	concurrency int

	// producer options
}

func defaultConsumerOptions() options {
	return options{
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
		brokers: []string{},
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

// WithTopics sets the topics to consume from.
func WithTopics(topics ...string) Option {
	return optionFunc(func(o *options) {
		o.topics = topics
	})
}

// WithBrokers sets the brokers to consume from.
func WithBrokers(brokers ...string) Option {
	return optionFunc(func(o *options) {
		o.brokers = brokers
	})
}

// WithKafkaConfig sets the kafka configmap values.
func WithKafkaConfig(key string, value interface{}) Option {
	return optionFunc(func(o *options) {
		_ = o.configMap.SetKey(key, value)
	})
}

// WithErrorHandler sets the error handler for the consumer.
func WithErrorHandler(errorHandler ErrorHandler) Option {
	return optionFunc(func(o *options) {
		o.errorHandler = errorHandler
	})
}

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
