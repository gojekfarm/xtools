package retry

import (
	"context"
	"errors"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/gojekfarm/xtools/xkafka"
)

// ErrRetryLimitExceeded is returned when the maximum number of retries is exceeded.
var ErrRetryLimitExceeded = errors.New("[xkafka/retry] retry limit exceeded")

// Option configures the retry middleware.
type Option interface {
	apply(*config)
}

// MaxRetries sets the maximum number of retries.
// ErrRetryLimitExceeded is returned after exhausting all retries.
type MaxRetries int

func (m MaxRetries) apply(c *config) { c.maxRetries = int(m) }

// MaxLifetime sets the maximum retry duration since the first execution.
// ErrRetryLimitExceeded is returned after the duration is exceeded.
type MaxLifetime time.Duration

func (m MaxLifetime) apply(c *config) { c.maxLifetime = time.Duration(m) }

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
	maxLifetime time.Duration
	delay       time.Duration
	jitter      time.Duration
	multiplier  float64
}

func newConfig(opts ...Option) *config {
	c := &config{}

	for _, opt := range opts {
		opt.apply(c)
	}

	return c
}

// ExponentialBackoff is a middleware with exponential backoff retry strategy.
func ExponentialBackoff(opts ...Option) xkafka.MiddlewareFunc {
	cfg := newConfig(opts...)

	return func(next xkafka.Handler) xkafka.Handler {
		return xkafka.HandlerFunc(func(ctx context.Context, msg *xkafka.Message) error {
			expBackoff := backoff.NewExponentialBackOff()
			expBackoff.InitialInterval = cfg.delay
			expBackoff.MaxElapsedTime = cfg.maxLifetime
			expBackoff.RandomizationFactor = float64(cfg.jitter) / float64(cfg.delay)
			expBackoff.Multiplier = cfg.multiplier

			attempt := 0

			return backoff.Retry(func() error {
				attempt++

				if attempt > cfg.maxRetries {
					return backoff.Permanent(ErrRetryLimitExceeded)
				}

				return next.Handle(ctx, msg)
			}, expBackoff)
		})
	}
}
