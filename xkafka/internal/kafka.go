// Package internal provides the kafka client interfaces for mocking.
package internal

import (
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

// ConsumerClient is the interface for the confluent-kafka-go/kafka.Consumer.
type ConsumerClient interface {
	GetMetadata(topic *string, allTopics bool, timeoutMs int) (*kafka.Metadata, error)
	ReadMessage(timeout time.Duration) (*kafka.Message, error)
	SubscribeTopics(topics []string, rebalanceCb kafka.RebalanceCb) error
	Unsubscribe() error
	Close() error
}

// ProducerClient is the interface for the confluent-kafka-go/kafka.Producer.
type ProducerClient interface {
	Produce(msg *kafka.Message, deliveryChan chan kafka.Event) error
	ProduceChannel() chan *kafka.Message
	Events() chan kafka.Event
	Flush(timeoutMs int) int
	Close()
}
