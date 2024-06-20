package xworker

import (
	"context"

	"github.com/sethvargo/go-retry"
)

// Enqueuer defines the enqueueing of a Job.
type Enqueuer interface {
	Enqueue(context.Context, *Job, ...Option) (*EnqueueResult, error)
}

// PeriodicEnqueuer defines the enqueueing of a periodic job.
type PeriodicEnqueuer interface {
	EnqueuePeriodically(cronSchedule string, job *Job, options ...Option) error
}

// EnqueuerFunc is helper for creating an Enqueuer.
type EnqueuerFunc func(context.Context, *Job, ...Option) (*EnqueueResult, error)

// PeriodicEnqueuerFunc is helper for creating a PeriodicEnqueuer.
type PeriodicEnqueuerFunc func(string, *Job, ...Option) error

// Enqueue enqueues a job with Option(s).
func (f EnqueuerFunc) Enqueue(ctx context.Context, j *Job, options ...Option) (*EnqueueResult, error) {
	return f(ctx, j, options...)
}

// EnqueuePeriodically schedules a Job with given cronSchedule and Option(s).
func (f PeriodicEnqueuerFunc) EnqueuePeriodically(cronSchedule string, job *Job, options ...Option) error {
	return f(cronSchedule, job, options...)
}

// EnqueueResult holds the data returned from worker implementation for further use.
type EnqueueResult struct {
	str string
	val interface{}
}

// NewEnqueueResult creates a new EnqueueResult.
func NewEnqueueResult(str string, val interface{}) *EnqueueResult {
	return &EnqueueResult{str: str, val: val}
}

// String returns a string representable value of EnqueueResult. It satisfies fmt.Stringer interface.
func (e *EnqueueResult) String() string {
	return e.str
}

// Value returns the concrete type returned by worker implementation, if any.
func (e *EnqueueResult) Value() interface{} {
	return e.val
}

// Enqueue enqueues a job with Option(s).
func (a *Adapter) Enqueue(ctx context.Context, job *Job, options ...Option) (*EnqueueResult, error) {
	return a.wrappedEnqueuer.Enqueue(ctx, job, options...)
}

// EnqueuePeriodically enqueues the job periodically governed by cron schedule.
func (a *Adapter) EnqueuePeriodically(cronSchedule string, job *Job, options ...Option) error {
	a.logger.Info().Fields(
		map[string]interface{}{
			"namespace":    a.namespace,
			"cronSchedule": cronSchedule,
			"jobName":      job.Name,
		},
	).Msg("EnqueuePeriodically")

	job.encoderFunc = a.encoderFunc

	return a.fulfiller.EnqueuePeriodically(cronSchedule, job, options...)
}

func (a *Adapter) enqueuer() Enqueuer {
	return EnqueuerFunc(func(ctx context.Context, j *Job, opts ...Option) (er *EnqueueResult, err error) {
		j.encoderFunc = a.encoderFunc

		if a.backoffStrategy == nil {
			return a.fulfiller.Enqueue(ctx, j, opts...)
		}

		err = retry.Do(ctx, a.backoffStrategy, func(ctx context.Context) error {
			var e error
			er, e = a.fulfiller.Enqueue(ctx, j, opts...)

			if e != nil {
				if a.onRetry != nil {
					a.onRetry(e)
				}

				return retry.RetryableError(e)
			}

			return nil
		})

		return
	})
}
