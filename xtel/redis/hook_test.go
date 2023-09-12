package redis

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
)

func TestDefaultOptions(t *testing.T) {
	tp := otel.GetTracerProvider()
	wtp := WithTracerProvider(tp)

	options := &options{}

	wtp(options)
	assert.Equal(t, options.tp, tp)
}

func TestWithTracerProvider(t *testing.T) {
	flag := false
	tp := providerFunc(func(name string, opts ...trace.TracerOption) trace.Tracer {
		flag = true
		return otel.GetTracerProvider().Tracer(name, opts...)
	})

	NewHook(WithTracerProvider(tp))
	assert.True(t, flag, "did not call custom TraceProvider")
}

type providerFunc func(name string, opts ...trace.TracerOption) trace.Tracer

func (fn providerFunc) Tracer(name string, opts ...trace.TracerOption) trace.Tracer {
	return fn(name, opts...)
}
