package xkafka

import (
	"context"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

// SimpleConsumer is a simple consumer that consumes messages from a Kafka topic.
type SimpleConsumer struct {
	name        string
	kafka       *kafka.Consumer
	middlewares []middleware
	config      consumerOptions
}

// NewSimpleConsumer creates a new SimpleConsumer instance.
func NewSimpleConsumer(name string, opts ...ConsumerOption) (*SimpleConsumer, error) {
	cfg := defaultConsumerOptions()

	for _, opt := range opts {
		opt(&cfg)
	}

	cfg.configMap.SetKey("bootstrap.servers", cfg.brokers)

	consumer, err := kafka.NewConsumer(&cfg.configMap)
	if err != nil {
		return nil, err
	}

	return &SimpleConsumer{
		name:   name,
		config: cfg,
		kafka:  consumer,
	}, nil
}

// GetMetadata returns the metadata for the consumer.
func (c *SimpleConsumer) GetMetadata() (*kafka.Metadata, error) {
	return c.kafka.GetMetadata(nil, false, int(c.config.metadataRequestTimeout.Milliseconds()))
}

// Use appends a MiddlewareFunc to the chain.
// Middleware can be used to intercept or otherwise modify, process or skip messages.
// They are executed in the order that they are applied to the Consumer.
func (c *SimpleConsumer) Use(mwf ...MiddlewareFunc) {
	for _, fn := range mwf {
		c.middlewares = append(c.middlewares, fn)
	}
}

// Start consumes from kafka and dispatches messages to handlers.
// It blocks until the context is cancelled or an error occurs.
// Errors are handled by the ErrorHandler if set, otherwise they are returned.
func (c *SimpleConsumer) Start(ctx context.Context, handler Handler) error {
	for i := len(c.middlewares) - 1; i >= 0; i-- {
		handler = c.middlewares[i].Middleware(handler)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			km, err := c.kafka.ReadMessage(c.config.pollTimeout)

			if err != nil {
				if kerr, ok := err.(kafka.Error); ok && kerr.Code() == kafka.ErrTimedOut {
					continue
				}

				if ferr := c.config.errorHandler(err); ferr != nil {
					return ferr
				}
			}

			msg := NewMessage(c.name, km)

			err = handler.Handle(ctx, msg)

			if ferr := c.config.errorHandler(err); ferr != nil {
				return ferr
			}
		}
	}
}

// Close closes the consumer.
func (c *SimpleConsumer) Close() error {
	<-time.After(c.config.shutdownTimeout)

	return c.kafka.Close()
}
