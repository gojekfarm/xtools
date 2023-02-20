package xkafka

import (
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type consumerOptions struct {
	topics       []string
	brokers      []string
	configMap    kafka.ConfigMap
	errorHandler ErrorHandler

	metadataRequestTimeout time.Duration
	pollTimeout            time.Duration
	shutdownTimeout        time.Duration

	concurrency int
}

func defaultConsumerOptions() consumerOptions {
	return consumerOptions{
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

// ConsumerOption is an interface for consumer options.
type ConsumerOption interface{ apply(*consumerOptions) }

type optionFunc func(*consumerOptions)

func (f optionFunc) apply(o *consumerOptions) { f(o) }

// WithTopics sets the topics to consume from.
func WithTopics(topics ...string) ConsumerOption {
	return optionFunc(func(o *consumerOptions) {
		o.topics = topics
	})
}

// WithBrokers sets the brokers to consume from.
func WithBrokers(brokers ...string) ConsumerOption {
	return optionFunc(func(o *consumerOptions) {
		o.brokers = brokers
	})
}

// WithKafkaConfig sets the kafka configmap values.
func WithKafkaConfig(key string, value interface{}) ConsumerOption {
	return optionFunc(func(o *consumerOptions) {
		o.configMap.SetKey(key, value)
	})
}

// WithErrorHandler sets the error handler for the consumer.
func WithErrorHandler(errorHandler ErrorHandler) ConsumerOption {
	return optionFunc(func(o *consumerOptions) {
		o.errorHandler = errorHandler
	})
}

// WithMetadataRequestTimeout sets the metadata request timeout.
func WithMetadataRequestTimeout(timeout time.Duration) ConsumerOption {
	return optionFunc(func(o *consumerOptions) {
		o.metadataRequestTimeout = timeout
	})
}

// WithPollTimeout sets the poll timeout.
func WithPollTimeout(timeout time.Duration) ConsumerOption {
	return optionFunc(func(o *consumerOptions) {
		o.pollTimeout = timeout
	})
}

// WithShutdownTimeout sets the shutdown timeout.
func WithShutdownTimeout(timeout time.Duration) ConsumerOption {
	return optionFunc(func(o *consumerOptions) {
		o.shutdownTimeout = timeout
	})
}

// WithConcurrency sets the concurrency level.
// Used with NewConcurrentConsumer only.
func WithConcurrency(concurrency int) ConsumerOption {
	return optionFunc(func(o *consumerOptions) {
		o.concurrency = concurrency
	})
}
