package xkafka

import "errors"

var (
	// ErrRetryable is the error message for retryable errors.
	ErrRetryable = errors.New("xkafka: retryable error")
	// ErrRequiredOption is returned when a required option is
	// not provided.
	ErrRequiredOption = errors.New("xkafka: required option not provided")
)

// ErrorHandler is a callback function that is called when an error occurs.
type ErrorHandler func(err error) error

func (h ErrorHandler) setConsumerConfig(o *consumerConfig) { o.errorHandler = h }

func (h ErrorHandler) setProducerConfig(o *producerConfig) { o.errorHandler = h }

// NoopErrorHandler is an ErrorHandler that passes the error through.
func NoopErrorHandler(err error) error { return err }
