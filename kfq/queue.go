package kfq

import (
	"context"
	"time"

	"github.com/rs/xid"
	"github.com/sethvargo/go-retry"

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
	retry retry.Backoff

	// retry config
	maxRetries  int
	delay       time.Duration
	jitter      time.Duration
	maxDuration time.Duration

	// error handling
	isErrorRetriable IsErrorRetriable
}

// NewQueue creates a new queue.
func NewQueue(name string, opts ...Option) (*Queue, error) {
	q := &Queue{
		name:             name,
		maxRetries:       10,
		delay:            200 * time.Millisecond,
		jitter:           50 * time.Millisecond,
		maxDuration:      5 * time.Minute,
		isErrorRetriable: isErrRetriable,
	}

	for _, opt := range opts {
		opt.apply(q)
	}

	expBackoff := retry.NewExponential(q.delay)
	expBackoff = retry.WithJitter(q.jitter, expBackoff)
	expBackoff = retry.WithMaxDuration(q.maxDuration, expBackoff)
	expBackoff = retry.WithMaxRetries(uint64(q.maxRetries), expBackoff)

	q.retry = expBackoff

	return q, nil
}

// Middleware returns a middleware which handles message lifecycle.
func (q *Queue) Middleware() xkafka.MiddlewareFunc {
	return func(next xkafka.Handler) xkafka.Handler {
		return xkafka.HandlerFunc(func(ctx context.Context, msg *xkafka.Message) error {
			ensureID(msg)

			if err := q.store.Put(ctx, msg); err != nil {
				return err
			}

			err := retry.Do(ctx, q.retry, func(ctx context.Context) error {
				err := next.Handle(ctx, msg)
				if err != nil && q.isErrorRetriable(err) {
					return retry.RetryableError(err)
				} else if err != nil {
					return err
				}

				return nil
			})
			if err != nil {
				return err
			}

			_ = q.store.Delete(ctx, msg.ID)

			return nil
		})
	}
}

// Run manages the lifecycle of the queue.
func (q *Queue) Run(ctx context.Context) error {
	defer q.Close()

	return q.Start(ctx)
}

// Start starts the queue.
func (q *Queue) Start(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		}
	}
}

// Close closes the queue.
func (q *Queue) Close() {

}

func ensureID(msg *xkafka.Message) {
	if msg.ID == "" {
		msg.ID = xid.New().String()
	}
}
