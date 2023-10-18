package grpc

import (
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// Option represents an option that can be used to configure the tracer provider.
type Option func(*options)

// WithTracerProvider returns an Option using the options type and TracerProvider.
// It sets the options for the given TracerProvider, which can be used for the
// instrumentation of Tracers.
func WithTracerProvider(trpr trace.TracerProvider) Option {
	return func(o *options) {
		o.tp = trpr
	}
}

// WithTextMapPropagator returns an Option using the options type and TextMapPropagator.
// It sets the options for the given TextMapPropagator.
// propagation.TextMapPropagator propagates cross-cutting concerns as key-value text pairs within a
// carrier that travels in-band across process boundaries.
func WithTextMapPropagator(propagator propagation.TextMapPropagator) Option {
	return func(o *options) {
		o.tmp = propagator
	}
}

type options struct {
	tp  trace.TracerProvider
	mp  metric.MeterProvider
	tmp propagation.TextMapPropagator
}
