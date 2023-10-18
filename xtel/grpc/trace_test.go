package grpc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultTracer(t *testing.T) {
	assert.Equal(t, defaultTracer, DefaultTracer())
}

func TestNewTracer(t *testing.T) {
	x := NewTracer()
	assert.NotNil(t, x.uci)
	assert.NotNil(t, x.usi)
}
