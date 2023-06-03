package prometheus

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestRegisterConsumerMetrics(t *testing.T) {
	r := prometheus.NewRegistry()

	assert.NoError(t, RegisterConsumerMetrics(r))
	assert.Error(t, RegisterConsumerMetrics(r))
}

func TestRegisterProducerMetrics(t *testing.T) {
	r := prometheus.NewRegistry()

	assert.NoError(t, RegisterProducerMetrics(r))
	assert.Error(t, RegisterProducerMetrics(r))
}
