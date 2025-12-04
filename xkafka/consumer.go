package xkafka

import (
	"context"
	"errors"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/sourcegraph/conc/stream"
)

// Consumer manages the consumption of messages from kafka topics
// and the processing of those messages.
type Consumer struct {
	name        string
	kafka       consumerClient
	handler     Handler
	middlewares []Middlewarer
	config      *consumerConfig
	cancelCtx   atomic.Pointer[context.CancelFunc]
	stopOffset  atomic.Bool

	// partition tracking
	mu               sync.Mutex
	activePartitions map[string]map[int32]struct{}
}

// NewConsumer creates a new Consumer instance.
func NewConsumer(name string, handler Handler, opts ...ConsumerOption) (*Consumer, error) {
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

	return &Consumer{
		name:             name,
		config:           cfg,
		kafka:            consumer,
		handler:          handler,
		activePartitions: make(map[string]map[int32]struct{}),
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

// Run starts running the Consumer. The component will stop running
// when the context is closed. Run blocks until the context is closed or
// an error occurs.
func (c *Consumer) Run(ctx context.Context) error {
	if err := c.subscribe(); err != nil {
		return err
	}

	if err := c.start(ctx); err != nil {
		return err
	}

	return c.close()
}

// Start subscribes to the configured topics and starts consuming messages.
// It runs the handler for each message in a separate goroutine.
//
// This method is non-blocking and returns immediately post subscribe.
// Instead, use Run if you want to block until the context is closed or an error occurs.
//
// Errors are handled by the ErrorHandler if set, otherwise they stop the consumer
// and are returned.
func (c *Consumer) Start() error {
	if err := c.subscribe(); err != nil {
		return err
	}

	ctx, cancel := context.WithCancel(context.Background())
	c.cancelCtx.Store(&cancel)

	go func() { _ = c.start(ctx) }()

	return nil
}

// Close closes the consumer.
func (c *Consumer) Close() {
	cancel := c.cancelCtx.Load()
	if cancel != nil {
		(*cancel)()
	}

	_ = c.close()
}

func (c *Consumer) start(ctx context.Context) error {
	c.handler = c.concatMiddlewares(c.handler)

	if c.config.concurrency > 1 {
		return c.runAsync(ctx)
	}

	return c.runSequential(ctx)
}

func (c *Consumer) runSequential(ctx context.Context) (err error) {
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

			err = c.handler.Handle(ctx, msg)
			if ferr := c.config.errorHandler(err); ferr != nil {
				err = ferr

				return err
			}

			err = c.storeMessage(msg)
			if err != nil {
				return err
			}
		}
	}
}

func (c *Consumer) runAsync(ctx context.Context) error {
	st := stream.New().WithMaxGoroutines(c.config.concurrency)
	ctx, cancel := context.WithCancelCause(ctx)

	for {
		select {
		case <-ctx.Done():
			st.Wait()

			defer cancel(nil)

			uerr := c.unsubscribe()

			err := context.Cause(ctx)
			if err != nil && errors.Is(err, context.Canceled) {
				return uerr
			}

			return errors.Join(err, uerr)
		default:
			km, err := c.kafka.ReadMessage(c.config.pollTimeout)
			if err != nil {
				var kerr kafka.Error
				if errors.As(err, &kerr) && kerr.Code() == kafka.ErrTimedOut {
					continue
				}

				if ferr := c.config.errorHandler(err); ferr != nil {
					cancel(ferr)

					continue
				}

				continue
			}

			msg := newMessage(c.name, km)

			st.Go(func() stream.Callback {
				err := c.handler.Handle(ctx, msg)
				if ferr := c.config.errorHandler(err); ferr != nil {
					cancel(ferr)

					return func() {
						c.stopOffset.Store(true)
					}
				}

				return func() {
					if err := c.storeMessage(msg); err != nil {
						cancel(err)
					}
				}
			})
		}
	}
}

func (c *Consumer) storeMessage(msg *Message) error {
	if msg.Status != Success && msg.Status != Skip {
		return nil
	}

	if c.stopOffset.Load() {
		return nil
	}

	// only store offset if partition is still active
	if !c.isPartitionActive(msg.Topic, msg.Partition) {
		return nil
	}

	// similar to StoreMessage in confluent-kafka-go/consumer.go
	// msg.Offset + 1 it ensures that the consumer starts with
	// next message when it restarts
	_, err := c.kafka.StoreOffsets([]kafka.TopicPartition{
		{
			Topic:     &msg.Topic,
			Partition: msg.Partition,
			Offset:    kafka.Offset(msg.Offset + 1),
		},
	})
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

func (c *Consumer) concatMiddlewares(h Handler) Handler {
	for i := len(c.middlewares) - 1; i >= 0; i-- {
		h = c.middlewares[i].Middleware(h)
	}

	return h
}

func (c *Consumer) subscribe() error {
	return c.kafka.SubscribeTopics(c.config.topics, c.rebalanceCallback)
}

func (c *Consumer) rebalanceCallback(_ *kafka.Consumer, event kafka.Event) error {
	switch e := event.(type) {
	case kafka.AssignedPartitions:
		c.onPartitionsAssigned(e.Partitions)
		return c.kafka.Assign(e.Partitions)

	case kafka.RevokedPartitions:
		if err := c.kafka.Unassign(); err != nil {
			return err
		}
		c.onPartitionsRevoked(e.Partitions)
	}

	return nil
}

func (c *Consumer) onPartitionsAssigned(partitions []kafka.TopicPartition) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, tp := range partitions {
		if tp.Topic == nil {
			continue
		}

		topic := *tp.Topic
		if c.activePartitions[topic] == nil {
			c.activePartitions[topic] = make(map[int32]struct{})
		}

		c.activePartitions[topic][tp.Partition] = struct{}{}
	}
}

func (c *Consumer) onPartitionsRevoked(partitions []kafka.TopicPartition) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, tp := range partitions {
		if tp.Topic == nil {
			continue
		}

		topic := *tp.Topic
		if c.activePartitions[topic] != nil {
			delete(c.activePartitions[topic], tp.Partition)

			if len(c.activePartitions[topic]) == 0 {
				delete(c.activePartitions, topic)
			}
		}
	}
}

func (c *Consumer) isPartitionActive(topic string, partition int32) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	// if no partitions tracked yet (before first rebalance), allow all
	if len(c.activePartitions) == 0 {
		return true
	}

	if partitions, ok := c.activePartitions[topic]; ok {
		_, active := partitions[partition]
		return active
	}

	return false
}

func (c *Consumer) unsubscribe() error {
	_, _ = c.kafka.Commit()

	return c.kafka.Unsubscribe()
}

func (c *Consumer) close() error {
	<-time.After(c.config.shutdownTimeout)

	return c.kafka.Close()
}
