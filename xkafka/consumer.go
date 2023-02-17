package xkafka

import (
	"context"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

// Consumer defines the interface for building a consumer.
type Consumer interface {
	GetMetadata() (*kafka.Metadata, error)
	Use(mwf ...MiddlewareFunc)
	Start(ctx context.Context, handler Handler) error
	Close() error
}

// Handler responds to a Message from a Kafka topic.
type Handler interface {
	Handle(ctx context.Context, m *Message) error
}

// HandlerFunc defines signature of a message handler function.
type HandlerFunc func(ctx context.Context, m *Message) error

// Handle implements Handler interface on HandlerFunc.
func (h HandlerFunc) Handle(ctx context.Context, m *Message) error {
	return h(ctx, m)
}

// MiddlewareFunc functions are closures that intercept Messages.
type MiddlewareFunc func(Handler) Handler

// middleware interface is anything which implements a MiddlewareFunc named Middleware.
type middleware interface {
	Middleware(handler Handler) Handler
}

// Middleware allows MiddlewareFunc to implement the middleware interface.
func (mw MiddlewareFunc) Middleware(handler Handler) Handler {
	return mw(handler)
}
