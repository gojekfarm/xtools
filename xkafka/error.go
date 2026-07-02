package xkafka

import (
	"errors"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

var (
	// ErrRetryable is the error message for retryable errors.
	ErrRetryable = errors.New("xkafka: retryable error")
	// ErrRequiredOption is returned when a required option is
	// not provided.
	ErrRequiredOption = errors.New("xkafka: required option not provided")
)

// isErrNoOffset reports whether err is librdkafka's ErrNoOffset
// ("Local: No offset stored"), which a manual Commit returns when there is
// nothing to commit for the current assignment. This is benign: with
// concurrency > 1 the poll loop can service a rebalance (Unassign clears the
// stored offsets) on another goroutine between StoreOffsets and Commit,
// leaving the store empty. The partition's new owner resumes from the last
// committed offset, so it is safe to treat as a no-op.
func isErrNoOffset(err error) bool {
	var kerr kafka.Error

	return errors.As(err, &kerr) && kerr.Code() == kafka.ErrNoOffset
}

// ErrorHandler is a callback function that is called when an error occurs.
type ErrorHandler func(err error) error

func (h ErrorHandler) setConsumerConfig(o *consumerConfig) { o.errorHandler = h }

func (h ErrorHandler) setProducerConfig(o *producerConfig) { o.errorHandler = h }

// NoopErrorHandler is an ErrorHandler that passes the error through.
func NoopErrorHandler(err error) error { return err }
