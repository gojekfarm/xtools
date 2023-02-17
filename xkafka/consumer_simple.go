package xkafka

import (
	"context"
	"time"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

var _ Consumer = &SimpleConsumer{}

// SimpleConsumer is a simple consumer that consumes messages from a Kafka topic.
type SimpleConsumer struct {
	config      ConsumerConfig
	kafka       *kafka.Consumer
	middlewares []middleware
}

// NewSimpleConsumer creates a new SimpleConsumer instance.
func NewSimpleConsumer(cfg ConsumerConfig) (*SimpleConsumer, error) {
	cfg.SetDefaults()

	kafkaCfg := &kafka.ConfigMap{
		"bootstrap.servers":           cfg.Brokers,
		"group.id":                    cfg.Group,
		"enable.auto.commit":          cfg.AutoCommit,
		"auto.offset.reset":           cfg.TopicOffset,
		"socket.keepalive.enable":     true,
		"session.timeout.ms":          int(cfg.SessionTimeout.Milliseconds()),
		"heartbeat.interval.ms":       int(cfg.HeartbeatInterval.Milliseconds()),
		"auto.commit.interval.ms":     int(cfg.AutoCommitInterval.Milliseconds()),
		"metadata.request.timeout.ms": cfg.MetadataRequestTimeout,
		"socket.timeout.ms":           int(cfg.SocketTimeout.Milliseconds()),
	}

	consumer, err := kafka.NewConsumer(kafkaCfg)
	if err != nil {
		return nil, err
	}

	err = consumer.SubscribeTopics(cfg.Topics, nil)
	if err != nil {
		return nil, err
	}

	return &SimpleConsumer{
		config: cfg,
		kafka:  consumer,
	}, nil
}

// GetMetadata is useful to verify that the client can connect successfully.
func (c *SimpleConsumer) GetMetadata() (*kafka.Metadata, error) {
	return c.kafka.GetMetadata(nil, false, c.config.MetadataRequestTimeout)
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
func (c *SimpleConsumer) Start(ctx context.Context, handler Handler) error {
	for i := len(c.middlewares) - 1; i >= 0; i-- {
		handler = c.middlewares[i].Middleware(handler)
	}

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			km, err := c.kafka.ReadMessage(c.config.PollTimeout)

			if err != nil {
				if kerr, ok := err.(kafka.Error); ok && kerr.Code() == kafka.ErrTimedOut {
					continue
				}

				return err
			}

			msg := NewMessage(c.config.Group, km)

			_ = handler.Handle(ctx, msg)
		}
	}
}

// Close closes the consumer.
func (c *SimpleConsumer) Close() error {
	<-time.After(c.config.ShutdownTimeout)

	return c.kafka.Close()
}
