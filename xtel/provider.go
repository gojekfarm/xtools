package xtel

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/hashicorp/go-multierror"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
)

// Provider helps manage the lifecycle and configuration of different Exporter(s)
// and other service level configurations for tracing and metrics.
type Provider struct {
	tp *trace.TracerProvider
	mp *metric.MeterProvider

	spanExporters []trace.SpanExporter
	metricReaders []metric.Reader

	roundTripper  http.RoundTripper
	samplingRatio float64

	initErrors *multierror.Error
}

// NewProvider creates a new Provider for the given ProviderOption.
func NewProvider(serviceName string, opts ...ProviderOption) (*Provider, error) {
	p := &Provider{
		roundTripper:  otelhttp.NewTransport(http.DefaultTransport),
		samplingRatio: 1,
	}

	for _, opt := range opts {
		opt.apply(p)
	}

	if err := p.initErrors.ErrorOrNil(); err != nil {
		return nil, err
	}

	p.initializeTraceProvider(serviceName)
	p.initializeMetricProvider(serviceName)

	return p, nil
}

// TracerProvider provides an OpenTelemetry TracerProvider over a given Provider
// which will be used to provide tracers to instrumentation.
func (p *Provider) TracerProvider() *TracerProvider { return p.tp }

// MeterProvider provides an OpenTelemetry MeterProvider over a given Provider
// which will be used to provide meters to instrumentation.
func (p *Provider) MeterProvider() *MeterProvider { return p.mp }

// Start starts Provider and initialises any related processes used for tracing.
func (p *Provider) Start() error {
	otel.SetTracerProvider(p.tp)
	otel.SetMeterProvider(p.mp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(&propagation.TraceContext{}))

	if p.roundTripper != nil {
		http.DefaultTransport = p.roundTripper
	}

	return nil
}

// Stop stops the Provider and closes any resources being used for tracing.
func (p *Provider) Stop() error {
	wg := &sync.WaitGroup{}
	errCh := make(chan error, 2)

	shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()

	for _, fn := range []func(context.Context) error{
		p.tp.Shutdown,
		p.mp.Shutdown,
	} {
		wg.Add(1)

		go func(closer func(context.Context) error) {
			defer wg.Done()

			if err := closer(shutdownCtx); err != nil {
				errCh <- err
			}
		}(fn)
	}

	wg.Wait()
	close(errCh)

	var errs *multierror.Error
	for err := range errCh {
		errs = multierror.Append(errs, err)
	}

	return errs.ErrorOrNil()
}

// Run will start running the Provider and associated processes.
// This method blocks until the passed context has been cancelled and then calls Stop.
// This makes Provider compatible with https://pkg.go.dev/github.com/gojekfarm/xrun package.
func (p *Provider) Run(ctx context.Context) error {
	// TODO: Revisit if p.Start() returns error
	_ = p.Start()

	<-ctx.Done()

	return p.Stop()
}

func (p *Provider) initializeTraceProvider(serviceName string) {
	traceOpts := []trace.TracerProviderOption{
		trace.WithResource(resource.NewSchemaless(semconv.ServiceNameKey.String(serviceName))),
		trace.WithSampler(trace.ParentBased(trace.TraceIDRatioBased(p.samplingRatio))),
	}

	for _, e := range p.spanExporters {
		traceOpts = append(traceOpts, trace.WithSpanProcessor(trace.NewBatchSpanProcessor(e)))
	}

	p.tp = trace.NewTracerProvider(traceOpts...)
}

func (p *Provider) initializeMetricProvider(serviceName string) {
	metricOpts := []metric.Option{
		metric.WithResource(resource.NewSchemaless(semconv.ServiceNameKey.String(serviceName))),
	}

	for _, r := range p.metricReaders {
		metricOpts = append(metricOpts, metric.WithReader(r))
	}

	p.mp = metric.NewMeterProvider(metricOpts...)
}
