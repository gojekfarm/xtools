package xworker

import (
	"context"
	"errors"
	"time"

	"github.com/rs/zerolog"
	"github.com/sethvargo/go-retry"
)

// Adapter is a utility to abstract away underlying worker.
type Adapter struct {
	namespace       string
	fulfiller       Fulfiller
	wrappedEnqueuer Enqueuer
	logger          zerolog.Logger

	backoffStrategy retry.Backoff
	onRetry         func(error)

	workerUpdateFetcher   WorkerUpdateFetcher
	metricCollector       MetricCollector
	metricCollectInterval time.Duration
	metricsCtx            context.Context
	metricsShutdown       context.CancelFunc

	handlerMiddlewares []handlerMiddleware
	enqueueMiddlewares []enqueueMiddleware

	encoderFunc PayloadEncoderFunc
	decoderFunc PayloadDecoderFunc
}

// NewAdapter creates an Adapter with provided AdapterOptions.
func NewAdapter(opts AdapterOptions) (*Adapter, error) {
	if opts.PayloadEncoder == nil {
		opts.PayloadEncoder = DefaultPayloadEncoderFunc
	}

	if opts.PayloadDecoder == nil {
		opts.PayloadDecoder = DefaultPayloadDecoderFunc
	}

	a := &Adapter{
		fulfiller:             opts.Fulfiller,
		logger:                opts.Logger,
		encoderFunc:           opts.PayloadEncoder,
		decoderFunc:           opts.PayloadDecoder,
		metricCollector:       opts.MetricCollector,
		metricCollectInterval: opts.MetricsCollectInterval,
		backoffStrategy:       opts.RetryBackoff,
		onRetry:               opts.OnRetryError,
	}

	a.wrappedEnqueuer = a.enqueuer()

	if opts.MetricCollector != nil {
		wuf, ok := opts.Fulfiller.(WorkerUpdateFetcher)
		if !ok {
			return nil, errors.New("AdapterOptions.Fulfiller must implement WorkerUpdateFetcher interface")
		}

		a.workerUpdateFetcher = wuf

		fods, ok := opts.Fulfiller.(FailOverDurationCollectorSetter)
		if ok {
			fods.SetFailOverDurationCollector(a.metricCollector.CollectFailOverDuration)
		}
	}

	a.metricsCtx, a.metricsShutdown = context.WithCancel(context.Background())

	return a, nil
}

// Start starts the Adapter and associated processes.
func (a *Adapter) Start() error {
	if err := a.fulfiller.Start(); err != nil {
		return err
	}

	if a.metricCollector != nil && a.workerUpdateFetcher != nil {
		go a.collectWorkerUpdatesPeriodically()
	}

	return nil
}

// Stop stops the Adapter and associated processes.
func (a *Adapter) Stop() error {
	a.metricsShutdown()

	return a.fulfiller.Stop()
}

// Run will start running the Provider and associated processes.
// This method blocks until the passed context has been cancelled and then calls Stop.
// This makes Adapter compatible with https://pkg.go.dev/github.com/gojekfarm/xrun package.
func (a *Adapter) Run(ctx context.Context) error {
	if err := a.Start(); err != nil {
		return err
	}

	<-ctx.Done()

	return a.Stop()
}
