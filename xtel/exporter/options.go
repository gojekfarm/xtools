package exporter

import (
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
)

// Option defines the implementation of the configurations.
type Option interface{ apply(*config) }

const defaultExporterEndpoint = "127.0.0.1:4317"

type config struct {
	mode                           Mode
	exporterEndpoint               string
	tracesExporterEndpointInsecure bool
	headers                        map[string]string
}

func newConfig(opts ...Option) *config {
	c := &config{
		headers:          make(map[string]string),
		exporterEndpoint: defaultExporterEndpoint,
	}

	for _, opt := range opts {
		opt.apply(c)
	}

	return c
}

func (c *config) grpcOptions() []otlptracegrpc.Option {
	res := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(c.exporterEndpoint),
		otlptracegrpc.WithHeaders(c.headers),
	}

	if c.tracesExporterEndpointInsecure {
		res = append(res, otlptracegrpc.WithInsecure())
	}

	return res
}

func (c *config) httpOptions() []otlptracehttp.Option {
	res := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(c.exporterEndpoint),
		otlptracehttp.WithHeaders(c.headers),
	}

	if c.tracesExporterEndpointInsecure {
		res = append(res, otlptracehttp.WithInsecure())
	}

	return res
}
