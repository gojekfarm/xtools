package redis

import (
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
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

// WithMeterProvider returns an Option using the options type and MeterProvider.
// It sets the options for the given MeterProvider, which can be used for the
// instrumentation of Meters.
func WithMeterProvider(mp metric.MeterProvider) Option {
	return func(o *options) {
		o.mp = mp
	}
}

// WithAttributes specifies additional attributes to be added to the span.
func WithAttributes(attrs ...attribute.KeyValue) Option {
	return func(o *options) {
		o.at = append(o.at, attrs...)
	}
}

type options struct {
	tp trace.TracerProvider
	mp metric.MeterProvider
	at []attribute.KeyValue
}

func newOptions(opts ...Option) *options {
	o := &options{
		tp: otel.GetTracerProvider(),
	}

	for _, opt := range opts {
		opt(o)
	}

	return o
}
