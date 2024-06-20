package xtel

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func TestEmptyTraceProvider(t *testing.T) {
	p := Provider{}
	assert.NotNil(t, p)
}

func TestNewProvider(t *testing.T) {
	np, err := NewProvider("dummy", DisableClientAutoTracing, SamplingFraction(0.1))
	assert.NotNil(t, np)
	assert.NoError(t, err)

	ctx, cancel := context.WithCancel(context.TODO())
	errCh := make(chan error, 1)

	go func() {
		errCh <- np.Run(ctx)
	}()

	cancel()

	assert.NoError(t, <-errCh)
}

func TestNewProvider_SpanExporter(t *testing.T) {
	tef := TraceExporterFunc(func(_ context.Context) (trace.SpanExporter, error) {
		return tracetest.NewInMemoryExporter(), nil
	})

	np, err := NewProvider("dummy", tef)

	if err != nil {
		panic(err)
	}
	assert.NotNil(t, np)
	assert.NoError(t, err)
	if err := np.Start(); err != nil {
		panic(err)
	}

	a := np.TracerProvider()
	assert.NotNil(t, a)

	assert.NoError(t, np.Stop())
}

func TestNewProvider_Reader(t *testing.T) {
	tef := MetricReaderFunc(func(_ context.Context) (metric.Reader, error) {
		return metric.NewManualReader(), nil
	})

	np, err := NewProvider("dummy", tef)

	if err != nil {
		panic(err)
	}
	assert.NotNil(t, np)
	assert.NoError(t, err)
	if err := np.Start(); err != nil {
		panic(err)
	}

	a := np.MeterProvider()
	assert.NotNil(t, a)

	assert.NoError(t, np.Stop())
}

func TestNewProvider_initError(t *testing.T) {
	np, err := NewProvider("dummy",
		TraceExporterFunc(func(_ context.Context) (trace.SpanExporter, error) {
			return nil, errors.New("can't create SpanExporter 1")
		}),
		MetricReaderFunc(func(_ context.Context) (metric.Reader, error) {
			return nil, errors.New("can't create Reader 2")
		}),
	)

	assert.Nil(t, np)
	assert.EqualError(t, err, `can't create SpanExporter 1
can't create Reader 2`)
}
