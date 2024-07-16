package xkafka

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/sourcegraph/conc/pool"
	"github.com/sourcegraph/conc/stream"
)

// BatchConsumer manages consumption & processing of messages
// from kafka topics in batches.
type BatchConsumer struct {
	name        string
	kafka       consumerClient
	handler     BatchHandler
	middlewares []BatchMiddlewarer
	config      *consumerConfig
	batch       *BatchManager
	cancelCtx   atomic.Pointer[context.CancelFunc]
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
		batch:   NewBatchManager(cfg.batchSize, cfg.batchTimeout),
	}, nil
}

// GetMetadata returns the metadata for the consumer.
func (c *BatchConsumer) GetMetadata() (*Metadata, error) {
	return c.kafka.GetMetadata(nil, false, int(c.config.metadataTimeout.Milliseconds()))
}

// Use appends a BatchMiddleware to the chain.
func (c *BatchConsumer) Use(mws ...BatchMiddlewarer) {
	c.middlewares = append(c.middlewares, mws...)
}

// Run starts the consumer and blocks until context is cancelled.
func (c *BatchConsumer) Run(ctx context.Context) error {
	if err := c.subscribe(); err != nil {
		return err
	}

	if err := c.start(ctx); err != nil {
		return err
	}

	return c.close()
}

// Start subscribes to the configured topics and starts consuming messages.
// This method is non-blocking and returns immediately post subscribe.
// Instead, use Run if you want to block until the context is closed or an error occurs.
//
// Errors are handled by the ErrorHandler if set, otherwise they stop the consumer
// and are returned.
func (c *BatchConsumer) Start() error {
	if err := c.subscribe(); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	c.cancelCtx.Store(&cancel)

	go func() { _ = c.start(ctx) }()

	return nil
}

// Close closes the consumer.
func (c *BatchConsumer) Close() {
	cancel := c.cancelCtx.Load()
	if cancel != nil {
		(*cancel)()
	}

	_ = c.close()
}

func (c *BatchConsumer) start(ctx context.Context) error {
	c.handler = c.concatMiddlewares(c.handler)

	pool := pool.New().
		WithContext(ctx).
		WithMaxGoroutines(2).
		WithCancelOnError()

	pool.Go(func(ctx context.Context) error {
		if c.config.concurrency > 1 {
			return c.processAsync(ctx)
		}

		return c.process(ctx)
	})

	pool.Go(func(ctx context.Context) error {
		return c.consume(ctx)
	})

	return pool.Wait()
}

func (c *BatchConsumer) process(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		case batch := <-c.batch.Receive():
			err := c.handler.HandleBatch(ctx, batch)
			if ferr := c.config.errorHandler(err); ferr != nil {
				return ferr
			}

			err = c.saveOffset(batch)
			if err != nil {
				return err
			}
		}
	}
}

func (c *BatchConsumer) processAsync(ctx context.Context) error {
	st := stream.New().WithMaxGoroutines(c.config.concurrency)
	ctx, cancel := context.WithCancelCause(ctx)

	for {
		select {
		case <-ctx.Done():
			st.Wait()

			err := context.Cause(ctx)
			if errors.Is(err, context.Canceled) {
				return nil
			}

			return err
		case batch := <-c.batch.Receive():
			st.Go(func() stream.Callback {
				err := c.handler.HandleBatch(ctx, batch)
				if ferr := c.config.errorHandler(err); ferr != nil {
					cancel(ferr)

					return func() {}
				}

				return func() {
					if err := c.saveOffset(batch); err != nil {
						cancel(err)
					}
				}
			})
		}
	}
}

func (c *BatchConsumer) consume(ctx context.Context) (err error) {
	if err := c.subscribe(); err != nil {
		return err
	}

	defer func() {
		if uerr := c.unsubscribe(); uerr != nil {
			err = errors.Join(err, uerr)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return err
		default:
			km, err := c.kafka.ReadMessage(c.config.pollTimeout)
			if err != nil {
				var kerr kafka.Error
				if errors.As(err, &kerr) && kerr.Code() == kafka.ErrTimedOut {
					continue
				}

				if ferr := c.config.errorHandler(err); ferr != nil {
					err = ferr

					return err
				}

				continue
			}

			msg := newMessage(c.name, km)

			c.batch.Add(msg)
		}
	}
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

func (c *BatchConsumer) concatMiddlewares(handler BatchHandler) BatchHandler {
	for i := len(c.middlewares) - 1; i >= 0; i-- {
		handler = c.middlewares[i].BatchMiddleware(handler)
	}

	return handler
}

func (c *BatchConsumer) saveOffset(batch *Batch) error {
	if batch.Status != Success && batch.Status != Skip {
		return nil
	}

	offsets := batch.GroupMaxOffset()

	_, err := c.kafka.StoreOffsets(offsets)
	if err != nil {
		return err
	}

	if c.config.manualCommit {
		if _, err := c.kafka.Commit(); err != nil {
			return err
		}
	}

	return nil
}

// BatchManager manages aggregation and processing of Message batches.
type BatchManager struct {
	size      int
	timeout   time.Duration
	batch     *Batch
	mutex     *sync.RWMutex
	flushChan chan *Batch
}

// NewBatchManager creates a new BatchManager.
func NewBatchManager(size int, timeout time.Duration) *BatchManager {
	b := &BatchManager{
		size:    size,
		timeout: timeout,
		mutex:   &sync.RWMutex{},
		batch: &Batch{
			Messages: make([]*Message, 0, size),
		},
		flushChan: make(chan *Batch),
	}

	go b.runFlushByTime()

	return b
}

// Add adds to batch and flushes when MaxSize is reached.
func (b *BatchManager) Add(m *Message) {
	b.mutex.Lock()
	b.batch.Messages = append(b.batch.Messages, m)

	if len(b.batch.Messages) >= b.size {
		b.flush()
	}

	b.mutex.Unlock()
}

// Receive returns a channel to read batched Messages.
func (b *BatchManager) Receive() <-chan *Batch {
	return b.flushChan
}

func (b *BatchManager) runFlushByTime() {
	t := time.NewTicker(b.timeout)

	for range t.C {
		b.mutex.Lock()
		b.flush()
		b.mutex.Unlock()
	}
}

// flush sends the batch to the flush channel and resets the batch.
// DESIGN: flush does NOT acquire a mutex lock. Locks should be managed by caller.
// nolint:gosimple
func (b *BatchManager) flush() {
	if len(b.batch.Messages) == 0 {
		return
	}

	b.flushChan <- b.batch

	b.batch = &Batch{
		Messages: make([]*Message, 0, b.size),
	}
}
