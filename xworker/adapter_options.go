package xworker

import (
	"time"

	"github.com/rs/zerolog"
	"github.com/sethvargo/go-retry"
)

// AdapterOptions holds the options for Adapter.
type AdapterOptions struct {
	// Fulfiller is an interface that defines what all behaviour is required from a worker implementation.
	Fulfiller Fulfiller
	Logger    zerolog.Logger
	// PayloadEncoderFunc is used to create an PayloadEncoder from io.Writer.
	PayloadEncoder PayloadEncoderFunc
	// PayloadDecoderFunc is used to create an PayloadDecoder from io.Reader.
	PayloadDecoder PayloadDecoderFunc
	// MetricCollector helps in gathering WorkerUpdate metrics.
	MetricCollector MetricCollector
	// MetricsCollectInterval defines the interval upon which MetricCollector.CollectWorkerUpdate
	// is called periodically.
	MetricsCollectInterval time.Duration
	// RetryBackoff defines the backoff strategy when an error occurs on Enqueuer.Enqueue call.
	RetryBackoff retry.Backoff
	// OnRetryError is called when an error has occurred on Enqueuer.Enqueue call.
	OnRetryError func(err error)
}
