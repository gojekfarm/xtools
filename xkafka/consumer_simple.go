package xkafka

import (
	"context"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/sourcegraph/conc/stream"
)

// Consumer is a simple consumer that consumes messages from a Kafka topic.
type Consumer struct {
	name        string
	kafka       *kafka.Consumer
	middlewares []middleware
	config      consumerOptions
}

// NewConsumer creates a new Consumer instance.
func NewConsumer(name string, opts ...ConsumerOption) (*Consumer, error) {
	cfg := defaultConsumerOptions()

	for _, opt := range opts {
		opt.apply(&cfg)
	}

	_ = cfg.configMap.SetKey("bootstrap.servers", cfg.brokers)

	consumer, err := kafka.NewConsumer(&cfg.configMap)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		name:   name,
		config: cfg,
		kafka:  consumer,
	}, nil
}

// GetMetadata returns the metadata for the consumer.
func (c *Consumer) GetMetadata() (*kafka.Metadata, error) {
	return c.kafka.GetMetadata(nil, false, int(c.config.metadataRequestTimeout.Milliseconds()))
}

// Use appends a MiddlewareFunc to the chain.
// Middleware can be used to intercept or otherwise modify, process or skip messages.
// They are executed in the order that they are applied to the Consumer.
func (c *Consumer) Use(mwf ...MiddlewareFunc) {
	for _, fn := range mwf {
		c.middlewares = append(c.middlewares, fn)
	}
}

// Start consumes from kafka and dispatches messages to handlers.
// It blocks until the context is cancelled or an error occurs.
// Errors are handled by the ErrorHandler if set, otherwise they stop the consumer
// and are returned.
func (c *Consumer) Start(ctx context.Context, handler Handler) error {
	if err := c.subscribe(); err != nil {
		return err
	}

	handler = c.concatMiddlewares(handler)

	st := stream.New().WithMaxGoroutines(c.config.concurrency)
	errChan := make(chan error, 1)

	for {
		select {
		case <-ctx.Done():
			st.Wait()

			return nil
		case err := <-errChan:
			st.Wait()

			return err
		default:
			km, err := c.kafka.ReadMessage(c.config.pollTimeout)

			if err != nil {
				if kerr, ok := err.(kafka.Error); ok && kerr.Code() == kafka.ErrTimedOut {
					continue
				}

				if ferr := c.config.errorHandler(err); ferr != nil {
					errChan <- ferr

					continue
				}
			}

			st.Go(func() stream.Callback {
				c.runHandler(ctx, handler, km, errChan)

				return func() {}
			})
		}
	}
}

func (c *Consumer) runHandler(ctx context.Context, handler Handler, km *kafka.Message, errChan chan error) {
	msg := NewMessage(c.name, km)

	err := handler.Handle(ctx, msg)

	if ferr := c.config.errorHandler(err); ferr != nil {
		errChan <- ferr
	}
}

func (c *Consumer) concatMiddlewares(h Handler) Handler {
	for i := len(c.middlewares) - 1; i >= 0; i-- {
		h = c.middlewares[i].Middleware(h)
	}

	return h
}

func (c *Consumer) subscribe() error {
	return c.kafka.SubscribeTopics(c.config.topics, nil)
}

// Close closes the consumer.
func (c *Consumer) Close() error {
	<-time.After(c.config.shutdownTimeout)

	_ = c.kafka.Unsubscribe()

	return c.kafka.Close()
}
