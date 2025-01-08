// Package retry provides middlewares for retrying transient errors.
package retry

import (
	"context"
	"errors"
	"time"

	"github.com/cenkalti/backoff/v4"

	"github.com/gojekfarm/xtools/xkafka"
)

var (
	// ErrPermanent is returned when the error should not be retried.
	ErrPermanent = errors.New("[xkafka/retry] permanent error")
)

// Option configures the retry middleware.
type Option interface {
	apply(*config)
}

// MaxRetries sets the maximum number of retries.
// ErrRetryLimitExceeded is returned after exhausting all retries.
type MaxRetries int

func (m MaxRetries) apply(c *config) { c.maxRetries = int(m) }

// MaxDuration sets the maximum retry duration since the first execution.
// It should be less than `max.poll.interval.ms` kafka consumer config.
type MaxDuration time.Duration

func (m MaxDuration) apply(c *config) { c.maxDuration = time.Duration(m) }

// Delay sets the delay between each retry.
type Delay time.Duration

func (d Delay) apply(c *config) { c.delay = time.Duration(d) }

// Jitter adds randomness to the delay between each retry.
type Jitter time.Duration

func (j Jitter) apply(c *config) { c.jitter = time.Duration(j) }

// Multiplier sets the exponential backoff multiplier.
type Multiplier float64

func (m Multiplier) apply(c *config) { c.multiplier = float64(m) }

type config struct {
	maxRetries  int
	maxDuration time.Duration
	delay       time.Duration
	jitter      time.Duration
	multiplier  float64
}

func newConfig(opts ...Option) *config {
	c := &config{
		maxRetries: 100,
		// less than 5 minutes default max.poll.interval.ms
		maxDuration: 299 * time.Second,
		delay:       200 * time.Millisecond,
		jitter:      20 * time.Millisecond,
		multiplier:  1.5,
	}

	for _, opt := range opts {
		opt.apply(c)
	}

	return c
}

// ExponentialBackoff is a middleware with exponential backoff retry strategy.
// It retries the handler until the maximum number of retries or the maximum
// lifetime is reached.
// Default values:
// - MaxRetries: 100
// - MaxDuration: 5 minutes
// - Delay: 200 milliseconds
// - Jitter: 20 milliseconds
// - Multiplier: 1.5
func ExponentialBackoff(opts ...Option) xkafka.MiddlewareFunc {
	cfg := newConfig(opts...)

	return func(next xkafka.Handler) xkafka.Handler {
		return xkafka.HandlerFunc(func(ctx context.Context, msg *xkafka.Message) error {
			expBackoff := backoff.NewExponentialBackOff()
			expBackoff.InitialInterval = cfg.delay
			expBackoff.MaxElapsedTime = cfg.maxDuration
			expBackoff.RandomizationFactor = float64(cfg.jitter) / float64(cfg.delay)
			expBackoff.Multiplier = cfg.multiplier

			b := backoff.WithMaxRetries(expBackoff, uint64(cfg.maxRetries))
			b = backoff.WithContext(b, ctx)

			return backoff.Retry(func() error {
				err := next.Handle(ctx, msg)
				if errors.Is(err, ErrPermanent) {
					return backoff.Permanent(err)
				}

				return err
			}, b)
		})
	}
}

// BatchExponentialBackoff is a middleware with exponential backoff retry strategy
// for xkafka.BatchConsumer.
func BatchExponentialBackoff(opts ...Option) xkafka.BatchMiddlewareFunc {
	cfg := newConfig(opts...)

	return func(next xkafka.BatchHandler) xkafka.BatchHandler {
		return xkafka.BatchHandlerFunc(func(ctx context.Context, batch *xkafka.Batch) error {
			expBackoff := backoff.NewExponentialBackOff()
			expBackoff.InitialInterval = cfg.delay
			expBackoff.MaxElapsedTime = cfg.maxDuration
			expBackoff.RandomizationFactor = float64(cfg.jitter) / float64(cfg.delay)
			expBackoff.Multiplier = cfg.multiplier

			b := backoff.WithMaxRetries(expBackoff, uint64(cfg.maxRetries))
			b = backoff.WithContext(b, ctx)

			return backoff.Retry(func() error {
				return next.HandleBatch(ctx, batch)
			}, b)
		})
	}
}
