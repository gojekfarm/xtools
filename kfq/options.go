package kfq

import "time"

// Option is an interface for options.
type Option interface{ apply(*Queue) }

// MaxRetries is maximum number of retries for a message.
// The message is marked as failed after this many retries.
// Default is 10.
type MaxRetries int

func (mr MaxRetries) apply(q *Queue) { q.maxRetries = int(mr) }

// Delay is the delay between retries.
// Default is 200ms.
type Delay time.Duration

func (d Delay) apply(q *Queue) { q.delay = time.Duration(d) }

// Jitter is the jitter added to the delay to avoid thundering herd.
// Default is 50ms.
type Jitter time.Duration

func (j Jitter) apply(q *Queue) { q.jitter = time.Duration(j) }

// MaxDuration is the maximum duration for which a message is retried.
// Default is 5 minutes.
type MaxDuration time.Duration

func (md MaxDuration) apply(q *Queue) { q.maxDuration = time.Duration(md) }
