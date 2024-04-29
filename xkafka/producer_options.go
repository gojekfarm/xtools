package xkafka

import (
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

const (
	atleastLeaderAck = 1
	partitioner      = "consistent_random"
)

// ProducerOption is an interface for producer options.
type ProducerOption interface{ setProducerConfig(*producerConfig) }

type producerConfig struct {
	brokers         []string
	configMap       kafka.ConfigMap
	errorHandler    ErrorHandler
	shutdownTimeout time.Duration
	producerFn      producerFunc
	deliveryCb      DeliveryCallback
}

func newProducerConfig(opts ...ProducerOption) (*producerConfig, error) {
	cfg := &producerConfig{
		producerFn: defaultProducerFunc,
		brokers:    []string{},
		configMap: kafka.ConfigMap{
			"default.topic.config": kafka.ConfigMap{
				"acks":        atleastLeaderAck,
				"partitioner": partitioner,
			},
		},
		errorHandler:    NoopErrorHandler,
		shutdownTimeout: 1 * time.Second,
	}

	for _, opt := range opts {
		opt.setProducerConfig(cfg)
	}

	return cfg, nil
}

// DeliveryCallback is a callback function triggered for every published message.
// Works only for xkafka.Producer.
type DeliveryCallback AckFunc

func (d DeliveryCallback) setProducerConfig(o *producerConfig) {
	o.deliveryCb = d
}
