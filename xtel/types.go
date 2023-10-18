package xtel

import (
	"context"

	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
)

// TracerProvider is an OpenTelemetry TracerProvider. It provides Tracers to
// instrumentation, so it can trace operational flow through a system.
type TracerProvider = trace.TracerProvider

// MeterProvider handles the creation and coordination of Meters. All Meters
// created by a MeterProvider will be associated with the same Resource, have
// the same Views applied to them, and have their produced metric telemetry
// passed to the configured Readers.
type MeterProvider = metric.MeterProvider

// TraceExporterFunc is used to create a new trace.SpanExporter.
type TraceExporterFunc func(context.Context) (trace.SpanExporter, error)

// MetricReaderFunc is used to create a new metric.Reader.
type MetricReaderFunc func(context.Context) (metric.Reader, error)
