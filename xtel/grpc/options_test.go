package grpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
)

func TestWithTraceProvider(t *testing.T) {
	tp := otel.GetTracerProvider()
	wtp := WithTracerProvider(tp)

	options := &options{}
	wtp(options)
	assert.Equal(t, options.tp, tp)
}

func TestWithMetricProvider(t *testing.T) {
	mp := otel.GetMeterProvider()
	wmp := WithMetricProvider(mp)

	options := &options{}
	wmp(options)
	assert.Equal(t, options.mp, mp)
}

func TestWithTextMapPropagator(t *testing.T) {
	tmp := otel.GetTextMapPropagator()
	wtmp := WithTextMapPropagator(tmp)

	options := &options{}
	wtmp(options)
	assert.Equal(t, options.tmp, tmp)
}
