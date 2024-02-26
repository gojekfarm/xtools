package http

import (
	"net/http"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// PathWhitelistFunc is used to control the trace paths.
// It can be used to skip routes that need not be traced.
type PathWhitelistFunc func(*http.Request) bool

// Option specifies instrumentation configuration options.
type Option interface{ apply(*options) }

// WithTracerProvider specifies a tracer provider to use for creating a tracer.
// If none is specified, the global provider is used.
func WithTracerProvider(provider trace.TracerProvider) Option {
	return optionFunc(func(cfg *options) { cfg.traceProvider = provider })
}

// WithTextMapPropagator specifies a propagation.TextMapPropagator to propagates cross-cutting concerns
// as key-value text pairs within a carrier that travels in-band across process boundaries.
func WithTextMapPropagator(propagator propagation.TextMapPropagator) Option {
	return optionFunc(func(cfg *options) { cfg.propagator = propagator })
}

func (pwf PathWhitelistFunc) apply(cfg *options) { cfg.pathWhitelistFunc = pwf }

type optionFunc func(*options)

func (o optionFunc) apply(c *options) { o(c) }

type options struct {
	traceProvider     trace.TracerProvider
	propagator        propagation.TextMapPropagator
	pathWhitelistFunc PathWhitelistFunc
}

func newOptions(opts ...Option) *options {
	o := &options{
		traceProvider:     otel.GetTracerProvider(),
		propagator:        otel.GetTextMapPropagator(),
		pathWhitelistFunc: defaultPathWhitelistFunc,
	}

	for _, opt := range opts {
		opt.apply(o)
	}

	return o
}
