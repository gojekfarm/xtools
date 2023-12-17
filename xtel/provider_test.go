package xtel

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/trace"
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
		o := []stdouttrace.Option{stdouttrace.WithWriter(os.Stdout)}
		o = append(o, stdouttrace.WithPrettyPrint())

		return stdouttrace.New(o...)
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

func TestNewProvider_initError(t *testing.T) {
	np, err := NewProvider("dummy",
		TraceExporterFunc(func(_ context.Context) (trace.SpanExporter, error) {
			return nil, errors.New("can't create SpanExporter 1")
		}),
		TraceExporterFunc(func(_ context.Context) (trace.SpanExporter, error) {
			return nil, errors.New("can't create SpanExporter 2")
		}),
	)

	assert.Nil(t, np)
	assert.EqualError(t, err, `can't create SpanExporter 1
can't create SpanExporter 2`)
}
