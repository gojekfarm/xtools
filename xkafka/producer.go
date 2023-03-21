package xkafka

import (
	"context"
	"strings"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/pkg/errors"
)

// Producer manages the production of messages to kafka topics.
// It provides both synchronous and asynchronous publish methods
// and a channel to stream delivery events.
type Producer struct {
	config              options
	kafka               producerClient
	events              chan kafka.Event
	middlewares         []middleware
	wrappedPublish      Handler
	wrappedAsyncPublish Handler
}

// NewProducer creates a new Producer.
func NewProducer(name string, opts ...Option) (*Producer, error) {
	cfg := defaultProducerOptions()

	for _, opt := range opts {
		opt.apply(&cfg)
	}

	_ = cfg.configMap.SetKey("bootstrap.servers", strings.Join(cfg.brokers, ","))
	_ = cfg.configMap.SetKey("client.id", name)

	producer, err := cfg.producerFn(&cfg.configMap)
	if err != nil {
		return nil, err
	}

	p := &Producer{
		config: cfg,
		kafka:  producer,
		events: producer.Events(),
	}

	p.wrappedPublish = p.publish()
	p.wrappedAsyncPublish = p.asyncPublish()

	return p, nil
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

// Start starts the kafka event handling.
// It blocks until the context is cancelled.
func (p *Producer) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case e := <-p.events:
			_ = p.handleEvent(e)
		}
	}
}

// Run manages both starting and stopping the producer.
func (p *Producer) Run(ctx context.Context) error {
	defer p.Close()

	return p.Start(ctx)
}

// AsyncPublish sends messages to the kafka topic asyncronously.
func (p *Producer) AsyncPublish(ctx context.Context, msg *Message) error {
	return p.wrappedAsyncPublish.Handle(ctx, msg)
}

func (p *Producer) asyncPublish() HandlerFunc {
	return func(ctx context.Context, msg *Message) error {
		km := newKafkaMessage(msg)

		p.kafka.ProduceChannel() <- km

		return nil
	}
}

// Publish sends messages to the kafka topic synchronously.
// Returns error if the message cannot be enqueued or if there's a Kafka error.
func (p *Producer) Publish(ctx context.Context, msg *Message) error {
	return p.wrappedPublish.Handle(ctx, msg)
}

func (p *Producer) publish() HandlerFunc {
	return func(ctx context.Context, msg *Message) error {
		deliveryChan := make(chan kafka.Event)
		defer close(deliveryChan)

		m := newKafkaMessage(msg)

		err := p.kafka.Produce(m, deliveryChan)
		if err != nil {
			// This is an enqueue error and should be retried
			re := errors.Wrap(err, ErrRetryable.Error())

			msg.AckFail(re)

			if p.config.deliveryCb != nil {
				p.config.deliveryCb(msg)
			}

			return re
		}

		e := <-deliveryChan

		return p.handleEvent(e)
	}
}

// Close waits for all messages to be delivered and closes the producer.
func (p *Producer) Close() {
	p.kafka.Flush(int(p.config.shutdownTimeout.Milliseconds()))

	p.kafka.Close()
}

func (p *Producer) handleEvent(e kafka.Event) error {
	switch ev := e.(type) {
	case *kafka.Message:
		m, ok := ev.Opaque.(*Message)
		if !ok {
			return nil
		}

		if ev.TopicPartition.Error != nil {
			m.AckFail(ev.TopicPartition.Error)
		} else {
			m.AckSuccess()
		}

		if p.config.deliveryCb != nil {
			p.config.deliveryCb(m)
		}

		return ev.TopicPartition.Error
	}

	return nil
}

func newKafkaMessage(msg *Message) *kafka.Message {
	km := msg.asKafkaMessage()

	km.TopicPartition.Partition = kafka.PartitionAny
	km.TimestampType = kafka.TimestampCreateTime

	return km
}
