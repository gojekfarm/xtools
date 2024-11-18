package xkafka

import (
	"context"
	"errors"
	"strings"
	"sync/atomic"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/sourcegraph/conc/stream"
)

// BatchConsumer manages the consumption of messages from kafka topics
// and processes them in batches.
type BatchConsumer struct {
	name        string
	kafka       consumerClient
	handler     BatchHandler
	middlewares []BatchMiddlewarer
	config      *consumerConfig
	stopOffset  atomic.Bool
}

// NewBatchConsumer creates a new BatchConsumer instance.
func NewBatchConsumer(name string, handler BatchHandler, opts ...ConsumerOption) (*BatchConsumer, error) {
	cfg, err := newConsumerConfig(opts...)
	if err != nil {
		return nil, err
	}

	// override kafka configs
	_ = cfg.configMap.SetKey("enable.auto.offset.store", false)
	_ = cfg.configMap.SetKey("bootstrap.servers", strings.Join(cfg.brokers, ","))
	_ = cfg.configMap.SetKey("group.id", name)

	if cfg.manualCommit {
		_ = cfg.configMap.SetKey("enable.auto.commit", false)
	}

	consumer, err := cfg.consumerFn(&cfg.configMap)
	if err != nil {
		return nil, err
	}

	return &BatchConsumer{
		name:    name,
		config:  cfg,
		kafka:   consumer,
		handler: handler,
	}, nil
}

// Use appends a BatchMiddlewareFunc to the chain.
func (c *BatchConsumer) Use(mwf ...BatchMiddlewarer) {
	c.middlewares = append(c.middlewares, mwf...)
}

// Run starts running the BatchConsumer. The component will stop running
// when the context is closed. Run blocks until the context is closed or
// an error occurs.
func (c *BatchConsumer) Run(ctx context.Context) (err error) {
	if err := c.subscribe(); err != nil {
		return err
	}

	defer func() {
		if cerr := c.close(); cerr != nil {
			err = errors.Join(err, cerr)
		}
	}()

	return c.start(ctx)
}

func (c *BatchConsumer) start(ctx context.Context) error {
	c.handler = c.concatMiddlewares(c.handler)

	if c.config.concurrency > 1 {
		return c.runAsync(ctx)
	}

	return c.runSequential(ctx)
}

func (c *BatchConsumer) runSequential(ctx context.Context) (err error) {
	defer func() {
		if uerr := c.unsubscribe(); uerr != nil {
			err = errors.Join(err, uerr)
		}
	}()

	batch := NewBatch()
	timer := time.NewTimer(c.config.batchTimeout)

	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			if err := c.processBatch(ctx, batch); err != nil {
				return err
			}

			return nil

		case <-timer.C:
			if len(batch.Messages) > 0 {
				if err := c.processBatch(ctx, batch); err != nil {
					return err
				}

				batch = NewBatch()
			}

			timer.Reset(c.config.batchTimeout)

		default:
			km, err := c.kafka.ReadMessage(c.config.pollTimeout)
			if err != nil {
				var kerr kafka.Error
				if errors.As(err, &kerr) && kerr.Code() == kafka.ErrTimedOut {
					continue
				}

				if ferr := c.config.errorHandler(err); ferr != nil {
					return ferr
				}

				continue
			}

			msg := newMessage(c.name, km)
			batch.Messages = append(batch.Messages, msg)

			if len(batch.Messages) >= c.config.batchSize {
				if err := c.processBatch(ctx, batch); err != nil {
					return err
				}

				batch = NewBatch()

				timer.Reset(c.config.batchTimeout)
			}
		}
	}
}

func (c *BatchConsumer) runAsync(ctx context.Context) error {
	st := stream.New().WithMaxGoroutines(c.config.concurrency)
	ctx, cancel := context.WithCancelCause(ctx)

	batch := NewBatch()
	timer := time.NewTimer(c.config.batchTimeout)

	defer timer.Stop()

	for {
		select {
		case <-ctx.Done():
			st.Wait()

			err := c.processBatch(ctx, batch)
			uerr := c.unsubscribe()
			err = errors.Join(err, uerr)

			cerr := context.Cause(ctx)
			if cerr != nil && !errors.Is(cerr, context.Canceled) {
				err = errors.Join(err, cerr)
			}

			return err

		case <-timer.C:
			if len(batch.Messages) > 0 {
				c.processBatchAsync(ctx, batch, st, cancel)
				batch = NewBatch()
			}

			timer.Reset(c.config.batchTimeout)

		default:
			km, err := c.kafka.ReadMessage(c.config.pollTimeout)
			if err != nil {
				var kerr kafka.Error
				if errors.As(err, &kerr) && kerr.Code() == kafka.ErrTimedOut {
					continue
				}

				if ferr := c.config.errorHandler(err); ferr != nil {
					cancel(ferr)
				}

				continue
			}

			msg := newMessage(c.name, km)
			batch.Messages = append(batch.Messages, msg)

			if len(batch.Messages) >= c.config.batchSize {
				c.processBatchAsync(ctx, batch, st, cancel)
				batch = NewBatch()

				timer.Reset(c.config.batchTimeout)
			}
		}
	}
}

func (c *BatchConsumer) processBatch(ctx context.Context, batch *Batch) error {
	if len(batch.Messages) == 0 {
		return nil
	}

	err := c.handler.HandleBatch(ctx, batch)
	if ferr := c.config.errorHandler(err); ferr != nil {
		return ferr
	}

	return c.storeBatch(batch)
}

func (c *BatchConsumer) processBatchAsync(
	ctx context.Context,
	batch *Batch,
	st *stream.Stream,
	cancel context.CancelCauseFunc,
) {
	st.Go(func() stream.Callback {
		err := c.handler.HandleBatch(ctx, batch)
		if ferr := c.config.errorHandler(err); ferr != nil {
			cancel(ferr)

			return func() {
				c.stopOffset.Store(true)
			}
		}

		return func() {
			if err := c.storeBatch(batch); err != nil {
				cancel(err)
			}
		}
	})
}

func (c *BatchConsumer) storeBatch(batch *Batch) error {
	if batch.Status != Success && batch.Status != Skip {
		return nil
	}

	if c.stopOffset.Load() {
		return nil
	}

	tps := batch.GroupMaxOffset()

	_, err := c.kafka.StoreOffsets(tps)
	if err != nil {
		return err
	}

	if c.config.manualCommit {
		_, err := c.kafka.Commit()
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *BatchConsumer) concatMiddlewares(h BatchHandler) BatchHandler {
	for i := len(c.middlewares) - 1; i >= 0; i-- {
		h = c.middlewares[i].BatchMiddleware(h)
	}

	return h
}

func (c *BatchConsumer) subscribe() error {
	return c.kafka.SubscribeTopics(c.config.topics, nil)
}

func (c *BatchConsumer) unsubscribe() error {
	_, _ = c.kafka.Commit()

	return c.kafka.Unsubscribe()
}

func (c *BatchConsumer) close() error {
	<-time.After(c.config.shutdownTimeout)

	return c.kafka.Close()
}
