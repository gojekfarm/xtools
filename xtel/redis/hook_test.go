package redis

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel"
)

func TestDefaultOptions(t *testing.T) {
	tp := otel.GetTracerProvider()
	wtp := WithTracerProvider(tp)

	options := &options{}

	wtp(options)
	assert.Equal(t, options.tp, tp)
}
