package xkafka

const (
	// ErrRetryable is the error message for retryable errors.
	ErrRetryable = "xkafka: retryable error"
	// ErrNoHandler is the error message for when a handler is not set.
	ErrNoHandler = "xkafka: no handler set"
)

// ErrorHandler is a callback function that is called when an error occurs.
type ErrorHandler func(err error) error

func (h ErrorHandler) apply(o *options) {
	o.errorHandler = h
}

// NoopErrorHandler is an ErrorHandler that passes the error through.
func NoopErrorHandler(err error) error {
	return err
}
