package grpc

import (
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
)

var defaultTracer = NewTracer()

// Tracer is a type that is used to access the UnaryClientInterceptor and UnaryServerInterceptor respectively.
type Tracer struct {
	uci grpc.UnaryClientInterceptor
	usi grpc.UnaryServerInterceptor
}

// NewTracer is used for implementing the traces for our Tracer.
func NewTracer(opts ...Option) *Tracer {
	o := &options{}

	for _, opt := range opts {
		opt(o)
	}

	gOpts := []otelgrpc.Option{
		otelgrpc.WithTracerProvider(o.tp),
		otelgrpc.WithMeterProvider(o.mp),
		otelgrpc.WithPropagators(o.tmp),
	}

	return &Tracer{
		uci: otelgrpc.UnaryClientInterceptor(gOpts...),
		usi: otelgrpc.UnaryServerInterceptor(gOpts...),
	}
}

// DefaultTracer is the default Tracer used by the package-level functions.
var DefaultTracer = defaultTracer
