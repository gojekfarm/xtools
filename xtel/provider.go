package xtel

import (
	"context"
	"net/http"
	"time"

	"github.com/hashicorp/go-multierror"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
)

// Provider helps manage the lifecycle and configuration of different trace.SpanExporter(s)
// and other service level configurations for tracing.
type Provider struct {
	tp            *trace.TracerProvider
	spanExporters []trace.SpanExporter
	roundTripper  http.RoundTripper
	initErrors    *multierror.Error
	samplingRatio float64
}

// TracerProvider is an OpenTelemetry TracerProvider. It provides Tracers to
// instrumentation, so it can trace operational flow through a system.
type TracerProvider = trace.TracerProvider

// NewExporterFunc is used to create a new trace.SpanExporter.
type NewExporterFunc func() (trace.SpanExporter, error)

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

	traceOpts := []trace.TracerProviderOption{
		trace.WithResource(resource.NewSchemaless(semconv.ServiceNameKey.String(serviceName))),
		trace.WithSampler(trace.ParentBased(trace.TraceIDRatioBased(p.samplingRatio))),
	}

	for _, e := range p.spanExporters {
		traceOpts = append(traceOpts, trace.WithSpanProcessor(trace.NewBatchSpanProcessor(e)))
	}

	p.tp = trace.NewTracerProvider(traceOpts...)

	return p, nil
}

// TracerProvider provides an OpenTelemetry TracerProvider over a given Provider
// which will be used to provide tracers to instrumentation.
func (p *Provider) TracerProvider() *TracerProvider {
	return p.tp
}

// Start starts Provider and initialises any related processes used for tracing.
func (p *Provider) Start() error {
	otel.SetTracerProvider(p.tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(&propagation.TraceContext{}))

	if p.roundTripper != nil {
		http.DefaultTransport = p.roundTripper
	}

	return nil
}

// Stop stops the Provider and closes any resources being used for tracing.
func (p *Provider) Stop() error {
	if len(p.spanExporters) > 0 {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), time.Second*30)
		defer cancel()

		return p.tp.Shutdown(shutdownCtx)
	}

	return nil
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
