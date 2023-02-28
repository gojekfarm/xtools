package xkafka

// ErrRetryable is the error message for retryable errors.
const ErrRetryable = "retryable"

// ErrorHandler is a callback function that is called when an error occurs.
type ErrorHandler func(err error) error

func (h ErrorHandler) apply(o *options) {
	o.errorHandler = h
}

// NoopErrorHandler is an ErrorHandler that passes the error through.
func NoopErrorHandler(err error) error {
	return err
}
