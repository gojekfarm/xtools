package xkafka

import (
	"time"
)

// Brokers sets the kafka brokers.
type Brokers []string

func (b Brokers) setConsumerConfig(o *consumerConfig) { o.brokers = b }

func (b Brokers) setProducerConfig(o *producerConfig) { o.brokers = b }

// ConfigMap allows setting kafka configuration.
type ConfigMap map[string]any

func (cm ConfigMap) setConsumerConfig(o *consumerConfig) {
	for k, v := range cm {
		_ = o.configMap.SetKey(k, v)
	}
}

func (cm ConfigMap) setProducerConfig(o *producerConfig) {
	for k, v := range cm {
		_ = o.configMap.SetKey(k, v)
	}
}

// ShutdownTimeout defines the timeout for the consumer/producer to shutdown.
type ShutdownTimeout time.Duration

func (st ShutdownTimeout) setConsumerConfig(o *consumerConfig) {
	o.shutdownTimeout = time.Duration(st)
}

func (st ShutdownTimeout) setProducerConfig(o *producerConfig) {
	o.shutdownTimeout = time.Duration(st)
}
