package prometheus

import (
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestRegisterConsumerMetrics(t *testing.T) {
	r := prometheus.NewRegistry()

	err := RegisterConsumerMetrics(r)
	assert.NoError(t, err)
}

func TestRegisterProducerMetrics(t *testing.T) {
	r := prometheus.NewRegistry()

	err := RegisterProducerMetrics(r)
	assert.NoError(t, err)
}
