package xkafka

import (
	"context"
	"log/slog"
	"strings"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
	"github.com/pkg/errors"
	"github.com/sourcegraph/conc/stream"
)

// Consumer manages the consumption of messages from kafka topics
// and the processing of those messages.
type Consumer struct {
	name        string
	kafka       consumerClient
	handler     Handler
	middlewares []middleware
	config      options
}

// NewConsumer creates a new Consumer instance.
func NewConsumer(name string, handler Handler, opts ...Option) (*Consumer, error) {
	cfg := defaultConsumerOptions()

	for _, opt := range opts {
		opt.apply(&cfg)
	}

	_ = cfg.configMap.SetKey("bootstrap.servers", strings.Join(cfg.brokers, ","))
	_ = cfg.configMap.SetKey("group.id", name)

	if cfg.manualOffset {
		_ = cfg.configMap.SetKey("enable.auto.offset.store", false)
	}

	consumer, err := cfg.consumerFn(&cfg.configMap)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		name:    name,
		config:  cfg,
		kafka:   consumer,
		handler: handler,
	}, nil
}

// GetMetadata returns the metadata for the consumer.
func (c *Consumer) GetMetadata() (*Metadata, error) {
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

// Run manages starting and stopping the consumer.
func (c *Consumer) Run(ctx context.Context) error {
	defer c.Close()

	return c.Start(ctx)
}

// Start subscribes to the configured topics and starts consuming messages.
// It runs the handler for each message in a separate goroutine.
// It blocks until the context is cancelled or an error occurs.
// Errors are handled by the ErrorHandler if set, otherwise they stop the consumer
// and are returned.
func (c *Consumer) Start(ctx context.Context) error {
	if err := c.subscribe(); err != nil {
		return err
	}

	c.handler = c.concatMiddlewares(c.handler)

	if c.config.concurrency > 1 {
		return c.runAsync(ctx)
	}

	return c.runSequential(ctx)
}

func (c *Consumer) runSequential(ctx context.Context) error {
	errChan := make(chan error, 1)

	for {
		select {
		case <-ctx.Done():
			return c.unsubscribe()
		case err := <-errChan:
			uerr := c.unsubscribe()
			if uerr != nil {
				// TODO: use multierror
				return errors.Wrap(err, uerr.Error())
			}

			return err
		default:
			km, err := c.kafka.ReadMessage(c.config.pollTimeout)

			if err != nil {
				if kerr, ok := err.(kafka.Error); ok && kerr.Code() == kafka.ErrTimedOut {
					continue
				}

				if ferr := c.config.errorHandler(err); ferr != nil {
					errChan <- err

					continue
				}
			}

			msg := newMessage(c.name, km)

			err = c.handler.Handle(ctx, msg)
			if ferr := c.config.errorHandler(err); ferr != nil {
				errChan <- ferr

				continue
			}

			if c.config.manualOffset &&
				(msg.Status == Success || msg.Status == Skip) {
				_, err := c.kafka.StoreMessage(km)
				if err != nil {
					slog.Error("[XKAFKA] Failed to store message", "error", err)

					errChan <- err
				}
			}
		}
	}
}

func (c *Consumer) runAsync(ctx context.Context) error {
	errChan := make(chan error, 1)
	p := stream.New().
		WithMaxGoroutines(c.config.concurrency)

	for {
		select {
		case <-ctx.Done():
			err := c.unsubscribe()

			p.Wait()

			return err
		case err := <-errChan:
			uerr := c.unsubscribe()
			if uerr != nil {
				// TODO: use multierror
				return errors.Wrap(err, uerr.Error())
			}

			p.Wait()

			return err
		default:
			km, err := c.kafka.ReadMessage(c.config.pollTimeout)

			if err != nil {
				if kerr, ok := err.(kafka.Error); ok && kerr.Code() == kafka.ErrTimedOut {
					continue
				}

				if ferr := c.config.errorHandler(err); ferr != nil {
					errChan <- err

					continue
				}
			}

			p.Go(func() stream.Callback {
				msg := newMessage(c.name, km)

				err = c.handler.Handle(ctx, msg)
				if ferr := c.config.errorHandler(err); ferr != nil {
					errChan <- ferr
				}

				return func() {
					if c.config.manualOffset &&
						(msg.Status == Success || msg.Status == Skip) {
						_, err := c.kafka.StoreMessage(km)
						if err != nil {
							errChan <- err
						}
					}
				}
			})
		}
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

func (c *Consumer) unsubscribe() error {
	c.kafka.Commit()

	return c.kafka.Unsubscribe()
}

// Close closes the consumer.
func (c *Consumer) Close() {
	<-time.After(c.config.shutdownTimeout)

	_ = c.kafka.Close()
}
