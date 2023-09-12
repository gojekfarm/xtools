package http

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

// NewHandler wraps the passed handler, functioning like middleware,
// in a span named after the operation and with any provided Options.
func NewHandler(h http.Handler, name string, opts ...Option) http.Handler {
	opt := newOptions(opts...)

	httpOptions := []otelhttp.Option{
		otelhttp.WithTracerProvider(opt.traceProvider),
		otelhttp.WithPropagators(opt.propagator),
	}

	return otelhttp.NewHandler(h, name, httpOptions...)
}
