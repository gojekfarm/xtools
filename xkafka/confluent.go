package xkafka

import (
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type consumerClient interface {
	GetMetadata(topic *string, allTopics bool, timeoutMs int) (*kafka.Metadata, error)
	ReadMessage(timeout time.Duration) (*kafka.Message, error)
	SubscribeTopics(topics []string, rebalanceCb kafka.RebalanceCb) error
	Unsubscribe() error
	StoreOffsets(offsets []kafka.TopicPartition) ([]kafka.TopicPartition, error)
	Commit() ([]kafka.TopicPartition, error)
	Close() error
}

type consumerFunc func(cfg *kafka.ConfigMap) (consumerClient, error)

func (cf consumerFunc) setConsumerConfig(o *consumerConfig) { o.consumerFn = cf }

func defaultConsumerFunc(cfg *kafka.ConfigMap) (consumerClient, error) {
	return kafka.NewConsumer(cfg)
}

type producerClient interface {
	Produce(msg *kafka.Message, deliveryChan chan kafka.Event) error
	ProduceChannel() chan *kafka.Message
	Events() chan kafka.Event
	Flush(timeoutMs int) int
	Close()
}

type producerFunc func(cfg *kafka.ConfigMap) (producerClient, error)

func (pf producerFunc) setProducerConfig(o *producerConfig) { o.producerFn = pf }

func defaultProducerFunc(cfg *kafka.ConfigMap) (producerClient, error) {
	return kafka.NewProducer(cfg)
}
