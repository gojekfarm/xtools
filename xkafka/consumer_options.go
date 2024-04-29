package xkafka

import (
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

// ConsumerOption is an interface for consumer options.
type ConsumerOption interface{ setConsumerConfig(*consumerConfig) }

type consumerConfig struct {
	brokers         []string
	configMap       kafka.ConfigMap
	errorHandler    ErrorHandler
	shutdownTimeout time.Duration
	consumerFn      consumerFunc
	topics          []string
	metadataTimeout time.Duration
	pollTimeout     time.Duration
	concurrency     int
	manualCommit    bool
}

func newConsumerConfig(opts ...ConsumerOption) (*consumerConfig, error) {
	cfg := &consumerConfig{
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

	for _, opt := range opts {
		opt.setConsumerConfig(cfg)
	}

	return cfg, nil
}

// Topics sets the kafka topics to consume.
type Topics []string

func (t Topics) setConsumerConfig(o *consumerConfig) { o.topics = t }

// PollTimeout defines the timeout for the consumer read timeout.
type PollTimeout time.Duration

func (pt PollTimeout) setConsumerConfig(o *consumerConfig) {
	o.pollTimeout = time.Duration(pt)
}

// MetadataTimeout defines the timeout for the consumer metadata request.
type MetadataTimeout time.Duration

func (mt MetadataTimeout) setConsumerConfig(o *consumerConfig) {
	o.metadataTimeout = time.Duration(mt)
}

// Concurrency defines the concurrency of the consumer.
type Concurrency int

func (c Concurrency) setConsumerConfig(o *consumerConfig) {
	o.concurrency = int(c)
}

// ManualCommit disables the auto commit and calls the `Commit` after every
// message is marked as `Success` or `Skip` by the handler.
//
// Works only for xkafka.Consumer.
//
// WARNING: Using this option will increase the message processing time,
// because of the synchronous `Commit` for every message.
type ManualCommit bool

func (mc ManualCommit) setConsumerConfig(o *consumerConfig) {
	o.manualCommit = bool(mc)
}
