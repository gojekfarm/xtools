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
	}
}

// ConsumerOption enables consumer configuration.
type ConsumerOption func(*consumerOptions)

// WithTopics sets the topics for the consumer.
func WithTopics(topics ...string) ConsumerOption {
	return func(o *consumerOptions) {
		o.topics = topics
	}
}

// WithBrokers sets the brokers for the consumer.
func WithBrokers(brokers ...string) ConsumerOption {
	return func(o *consumerOptions) {
		o.brokers = brokers
	}
}

// WithKafkaConfig sets the Kafka config for the consumer.
func WithKafkaConfig(key string, value interface{}) ConsumerOption {
	return func(o *consumerOptions) {
		o.configMap.SetKey(key, value)
	}
}

// WithErrorHandler sets the error handler for the consumer.
func WithErrorHandler(fn ErrorHandler) ConsumerOption {
	return func(o *consumerOptions) {
		o.errorHandler = fn
	}
}

// WithMetadataRequestTimeout sets the timeout for metadata requests.
func WithMetadataRequestTimeout(timeout time.Duration) ConsumerOption {
	return func(o *consumerOptions) {
		o.metadataRequestTimeout = timeout
	}
}

// WithPollTimeout sets the timeout for polling messages.
func WithPollTimeout(timeout time.Duration) ConsumerOption {
	return func(o *consumerOptions) {
		o.pollTimeout = timeout
	}
}

// WithShutdownTimeout sets the timeout for shutting down the consumer.
func WithShutdownTimeout(timeout time.Duration) ConsumerOption {
	return func(o *consumerOptions) {
		o.shutdownTimeout = timeout
	}
}
