package xworker

import (
	"context"
	"time"
)

// RegisterOptions can be used to customise the Job Handler registration.
type RegisterOptions struct {
	// Priority of the queue for the Job.
	Priority int
	// MaxRetries, 0: send straight to dead/archive (unless SkipArchive)
	MaxRetries int
	// SkipArchive when true, will not send failed jobs to the dead queue/archival when retries are exhausted.
	SkipArchive bool
	// MaxConcurrency denotes max number of jobs to keep in flight (default is 0, meaning no max)
	MaxConcurrency uint
	// RetryBackoffStrategy can be used to calculate the retry delay duration for a failed task.
	RetryBackoffStrategy RetryBackoffStrategy
}

// Registerer defines the signature to register a Job Handler.
type Registerer interface {
	RegisterHandlerWithOptions(jobName string, jobHandler Handler, options RegisterOptions) error
}

// RegistererFunc is a helper to create Registerer.
type RegistererFunc func(jobName string, jobHandler Handler, options RegisterOptions) error

// RegisterHandlerWithOptions registers a Job Handler.
func (f RegistererFunc) RegisterHandlerWithOptions(jobName string, jobHandler Handler, options RegisterOptions) error {
	return f(jobName, jobHandler, options)
}

// RegisterHandlerWithOptions adds the HandlerFunc for a Job with given jobName and RegisterOptions.
func (a *Adapter) RegisterHandlerWithOptions(
	jobName string,
	jobHandler Handler,
	options RegisterOptions,
) error {
	h := a.wrapHandlerWithMiddlewares(jobHandler)

	a.injectRetryBackoffWithDecoder(&options)

	return a.fulfiller.RegisterHandlerWithOptions(jobName, HandlerFunc(func(ctx context.Context, j *Job) error {
		j.decoderFunc = a.decoderFunc

		return h.Handle(ctx, j)
	}), options)
}

func (a *Adapter) injectRetryBackoffWithDecoder(options *RegisterOptions) {
	if rbs := options.RetryBackoffStrategy; rbs != nil {
		options.RetryBackoffStrategy = RetryBackoffFunc(func(n int, err error, j *Job) time.Duration {
			j.decoderFunc = a.decoderFunc

			return rbs.RetryBackoff(n, err, j)
		})
	}
}
