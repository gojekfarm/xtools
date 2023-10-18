package exporter

import (
	"context"
	"os"

	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/trace"

	"github.com/gojekfarm/xtools/xtel"
)

// STDOutOptions handles the configuration of STDOUT output. It handles the PrettyPrint parameter of the application.
type STDOutOptions struct {
	PrettyPrint bool
}

// NewSTDOut exports and writes exported tracing telemetry information in JSON for the given STDOutOptions.
func NewSTDOut(opts STDOutOptions) xtel.TraceExporterFunc {
	return func(_ context.Context) (trace.SpanExporter, error) {
		o := []stdouttrace.Option{stdouttrace.WithWriter(os.Stdout)}

		if opts.PrettyPrint {
			o = append(o, stdouttrace.WithPrettyPrint())
		}

		return stdouttrace.New(o...)
	}
}
