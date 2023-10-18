package exporter

import (
	"context"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/trace"

	"github.com/gojekfarm/xtools/xtel"
)

// Mode is the used to control the export configuration of the otlptrace package exporters.
type Mode int

const (
	// GRPC is the default exporters configurations for the Exporter.
	GRPC Mode = iota
	// HTTP can set as the configuration for the Exporter.
	HTTP
)

// WithTracesExporterEndpoint configures the endpoint for sending traces via OTLP.
func WithTracesExporterEndpoint(url string) Option {
	return optionFunc(func(c *config) { c.exporterEndpoint = url })
}

// WithTracesExporterInsecure permits connecting to the
// trace endpoint without a certificate.
var WithTracesExporterInsecure = &insecure{}

// NewOTLP returns a new exporter and starts it. It configs the new exporter with the given Options.
func NewOTLP(opts ...Option) xtel.TraceExporterFunc {
	return func() (trace.SpanExporter, error) {
		cfg := newConfig(opts...)

		var client otlptrace.Client
		switch cfg.mode {
		case GRPC:
			client = otlptracegrpc.NewClient(cfg.grpcOptions()...)
		case HTTP:
			client = otlptracehttp.NewClient(cfg.httpOptions()...)
		}

		traceExporter, err := newTraceExporter(context.Background(), client)
		if err != nil {
			return nil, err
		}

		return traceExporter, nil
	}
}

var newTraceExporter = otlptrace.New

func (m Mode) apply(c *config) { c.mode = m }

type insecure struct {
}

func (i *insecure) apply(c *config) { c.tracesExporterEndpointInsecure = true }

type optionFunc func(*config)

func (f optionFunc) apply(c *config) { f(c) }
