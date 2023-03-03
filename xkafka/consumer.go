package xkafka

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/sourcegraph/conc/stream"
)

// KafkaConsumer is the interface for kafka consumer.
type KafkaConsumer interface {
	GetMetadata(topic *string, allTopics bool, timeoutMs int) (*kafka.Metadata, error)
	ReadMessage(timeout time.Duration) (*kafka.Message, error)
	SubscribeTopics(topics []string, rebalanceCb kafka.RebalanceCb) error
	Unsubscribe() error
	Close() error
}

// ConsumerFunc is a function that returns a KafkaConsumer.
type ConsumerFunc func(cfg *kafka.ConfigMap) (KafkaConsumer, error)

func (cf ConsumerFunc) apply(o *options) { o.consumerFn = cf }

// DefaultConsumerFunc is the default ConsumerFunc that initialises
// a new confluent-kafka-go/kafka.Consumer.
func DefaultConsumerFunc(cfg *kafka.ConfigMap) (KafkaConsumer, error) {
	return kafka.NewConsumer(cfg)
}

// Consumer manages the consumption of messages from kafka topics
// and the processing of those messages.
type Consumer struct {
	name        string
	kafka       KafkaConsumer
	handler     Handler
	middlewares []middleware
	config      options
}

// NewConsumer creates a new Consumer instance.
func NewConsumer(name string, opts ...Option) (*Consumer, error) {
	cfg := defaultConsumerOptions()

	for _, opt := range opts {
		opt.apply(&cfg)
	}

	_ = cfg.configMap.SetKey("bootstrap.servers", strings.Join(cfg.brokers, ","))

	consumer, err := cfg.consumerFn(&cfg.configMap)
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
	return c.kafka.GetMetadata(nil, false, int(c.config.metadataTimeout.Milliseconds()))
}

// Use appends a MiddlewareFunc to the chain.
// Middleware can be used to intercept or otherwise modify, process or skip messages.
// They are executed in the order that they are applied to the Consumer.
func (c *Consumer) Use(mwf ...MiddlewareFunc) {
	for _, fn := range mwf {
		c.middlewares = append(c.middlewares, fn)
	}
}

// WithHandler sets the message handler for the consumer.
// Any previously set handlers are overwritten.
func (c *Consumer) WithHandler(handler Handler) *Consumer {
	c.handler = handler

	return c
}

// Start consumes from kafka and dispatches messages to handlers.
// It blocks until the context is cancelled or an error occurs.
// Errors are handled by the ErrorHandler if set, otherwise they stop the consumer
// and are returned.
func (c *Consumer) Start(ctx context.Context) error {
	if c.handler == nil {
		return errors.New(ErrNoHandler)
	}

	if err := c.subscribe(); err != nil {
		return err
	}

	c.handler = c.concatMiddlewares(c.handler)

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
				c.runHandler(ctx, c.handler, km, errChan)

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
