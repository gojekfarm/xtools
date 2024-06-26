package badgerq

import "time"

// Option configures a BadgerQueue.
type Option interface {
	apply(*config)
}

// MaxRetries sets the maximum number of retries.
type MaxRetries int

func (m MaxRetries) apply(c *config) { c.maxRetries = int(m) }

// MaxDuration sets the maximum retry duration since the first execution.
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
		maxRetries:  100,
		maxDuration: time.Hour,
		delay:       200 * time.Millisecond,
		jitter:      20 * time.Millisecond,
		multiplier:  1.5,
	}

	for _, opt := range opts {
		opt.apply(c)
	}

	return c
}
