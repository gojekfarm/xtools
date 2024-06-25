package xworker

import "time"

// RetryBackoffStrategy can be used to calculate the retry delay duration for a failed task.
type RetryBackoffStrategy interface {
	// RetryBackoff returns the backoff duration when retrying a Job, given
	// the retry count(starting from 0), error, and the Job.
	RetryBackoff(retryCount int, err error, job *Job) time.Duration
}

// RetryBackoffFunc can be used to calculate the retry delay duration for a failed task given
// the retry count(starting from 0), error, and the Job.
type RetryBackoffFunc func(int, error, *Job) time.Duration

// RetryBackoff returns the backoff duration when retrying a Job, given
// the retry count(starting from 0), error, and the Job.
func (f RetryBackoffFunc) RetryBackoff(n int, err error, j *Job) time.Duration { return f(n, err, j) }

// ConstantRetryBackoff always returns same time.Duration when calling RetryBackoffStrategy.RetryBackoff.
type ConstantRetryBackoff time.Duration

// RetryBackoff will always return same time.Duration as ConstantRetryBackoff.
func (c ConstantRetryBackoff) RetryBackoff(_ int, _ error, _ *Job) time.Duration {
	return time.Duration(c)
}

// LinearRetryBackoff returns time.Duration values which will increase/decrease in Step(s) starting from InitialDelay.
type LinearRetryBackoff struct {
	InitialDelay, Step time.Duration
}

// RetryBackoff returns time.Duration values which will increase/decrease in Step(s) starting from InitialDelay.
func (l LinearRetryBackoff) RetryBackoff(n int, _ error, _ *Job) time.Duration {
	return l.InitialDelay + (l.Step * time.Duration(n))
}
