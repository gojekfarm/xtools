package kfq

import (
	"context"
	"time"

	"github.com/gojekfarm/xtools/xkafka"
)

// MessageStorer is an interface for storing messages.
type MessageStorer interface {
	Put(ctx context.Context, msg *xkafka.Message) error
	Get(ctx context.Context, id string) (*xkafka.Message, error)
	Delete(ctx context.Context, id string) error
	Iterator(ctx context.Context, fn func(*xkafka.Message)) error
}

// Queue ensures reliable execution of messages. It manages
// persistence and retry policies for messages until they are
// successfully processed.
type Queue struct {
	name  string
	store MessageStorer

	// retry config
	maxRetries  int
	delay       time.Duration
	jitter      time.Duration
	maxDuration time.Duration
}

// NewQueue creates a new queue.
func NewQueue(name string, opts ...Option) *Queue {
	q := &Queue{
		name:        name,
		maxRetries:  10,
		delay:       200 * time.Millisecond,
		jitter:      50 * time.Millisecond,
		maxDuration: 5 * time.Minute,
	}

	for _, opt := range opts {
		opt.apply(q)
	}

	return q
}

// Middleware returns a middleware which handles message lifecycle.
func (q *Queue) Middleware() xkafka.MiddlewareFunc {
	return func(next xkafka.Handler) xkafka.Handler {
		return xkafka.HandlerFunc(func(ctx context.Context, msg *xkafka.Message) error {
			// TODO: add retry logic

			return next.Handle(ctx, msg)
		})
	}
}

// Run manages the lifecycle of the queue.
func (q *Queue) Run(ctx context.Context) error {
	return nil
}
