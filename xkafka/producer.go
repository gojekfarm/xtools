package xkafka

import (
	"context"
	"strings"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/pkg/errors"
)

// KafkaProducer is the interface for kafka producer.
type KafkaProducer interface {
	Produce(msg *kafka.Message, deliveryChan chan kafka.Event) error
	ProduceChannel() chan *kafka.Message
	Events() chan kafka.Event
	Flush(timeoutMs int) int
	Close()
}

// ProducerFunc is a function that returns a KafkaProducer.
type ProducerFunc func(cfg *kafka.ConfigMap) (KafkaProducer, error)

func (pf ProducerFunc) apply(o *options) { o.producerFn = pf }

// DefaultProducerFunc is the default producer function that initializes
// a new confluent-kafka-go/kafka.Producer.
func DefaultProducerFunc(cfg *kafka.ConfigMap) (KafkaProducer, error) {
	return kafka.NewProducer(cfg)
}

// Producer manages the production of messages to kafka topics.
// It provides both synchronous and asynchronous publish methods
// and a channel to stream delivery events.
type Producer struct {
	config              options
	kafka               KafkaProducer
	delivery            chan *Message
	middlewares         []middleware
	wrappedPublish      Handler
	wrappedAsyncPublish Handler
}

// NewProducer creates a new Producer.
func NewProducer(opts ...Option) (*Producer, error) {
	cfg := defaultProducerOptions()

	for _, opt := range opts {
		opt.apply(&cfg)
	}

	_ = cfg.configMap.SetKey("bootstrap.servers", strings.Join(cfg.brokers, ","))

	producer, err := cfg.producerFn(&cfg.configMap)
	if err != nil {
		return nil, err
	}

	p := &Producer{
		config:   cfg,
		kafka:    producer,
		delivery: make(chan *Message),
	}

	p.start()

	return p, nil
}

// DeliveryEvents returns a read-only channel that streams acked messages.
func (p *Producer) DeliveryEvents() <-chan *Message {
	return p.delivery
}

// Use appends a MiddlewareFunc to the chain.
// Middleware can be used to intercept or otherwise modify, process or skip messages.
// They are executed in the order that they are applied to the Producer.
func (p *Producer) Use(mwf ...MiddlewareFunc) {
	for _, fn := range mwf {
		p.middlewares = append(p.middlewares, fn)
	}

	p.wrappedPublish = p.publish()

	for i := len(p.middlewares) - 1; i >= 0; i-- {
		p.wrappedPublish = p.middlewares[i].Middleware(p.wrappedPublish)
	}

	p.wrappedAsyncPublish = p.asyncPublish()

	for i := len(p.middlewares) - 1; i >= 0; i-- {
		p.wrappedAsyncPublish = p.middlewares[i].Middleware(p.wrappedAsyncPublish)
	}
}

// Start processes kafka delivery events and passes them over to delivery handler.
func (p *Producer) start() {
	go func() {
		for e := range p.kafka.Events() {
			switch ev := e.(type) {
			case *kafka.Message:
				m, ok := ev.Opaque.(*Message)
				if !ok {
					continue
				}

				if ev.TopicPartition.Error != nil {
					m.AckFail(ev.TopicPartition.Error)
				} else {
					m.AckSuccess()
				}

				// make a copy of the message for delivery channel
				p.delivery <- m.Copy()
			}
		}
	}()
}

// AsyncPublish sends messages to the kafka topic asyncronously.
func (p *Producer) AsyncPublish(ctx context.Context, msg *Message) error {
	return p.wrappedAsyncPublish.Handle(ctx, msg)
}

func (p *Producer) asyncPublish() HandlerFunc {
	return HandlerFunc(func(ctx context.Context, msg *Message) error {
		km := newKafkaMessage(msg)

		p.kafka.ProduceChannel() <- km

		return nil
	})
}

// Publish sends messages to the kafka topic synchronously.
// Returns error if the message cannot be enqueued or if there's a Kafka error.
func (p *Producer) Publish(ctx context.Context, msg *Message) error {
	return p.wrappedPublish.Handle(ctx, msg)
}

func (p *Producer) publish() HandlerFunc {
	return HandlerFunc(func(ctx context.Context, msg *Message) error {
		deliveryChan := make(chan kafka.Event)
		defer close(deliveryChan)

		m := newKafkaMessage(msg)

		err := p.kafka.Produce(m, deliveryChan)
		if err != nil {
			// This is an enqueue error and should be retried
			re := errors.Wrap(err, ErrRetryable)

			msg.AckFail(re)

			return re
		}

		e := <-deliveryChan

		switch ev := e.(type) {
		case *kafka.Message:
			msg.Partition = ev.TopicPartition.Partition

			if ev.TopicPartition.Error != nil {
				msg.AckFail(ev.TopicPartition.Error)

				return ev.TopicPartition.Error
			}

			msg.AckSuccess()

			// send a copy of the message to delivery channel
			// because the message is not sent to kafka.Events() channel
			// when using sync publish
			p.delivery <- msg.Copy()
		}

		return nil
	})
}

// Close closes delivery & ack channels and the underlying Kafka client.
func (p *Producer) Close(ctx context.Context) {
	p.kafka.Flush(int(p.config.shutdownTimeout * time.Millisecond))

	// close kafka client
	p.kafka.Close()

	// and then close delivery channel
	close(p.delivery)
}

func newKafkaMessage(msg *Message) *kafka.Message {
	return &kafka.Message{
		TopicPartition: kafka.TopicPartition{
			Topic:     &msg.Topic,
			Partition: kafka.PartitionAny,
		},
		Key:           msg.Key,
		Value:         msg.Value,
		Timestamp:     msg.Timestamp,
		TimestampType: kafka.TimestampCreateTime,
		Opaque:        msg,
	}
}
