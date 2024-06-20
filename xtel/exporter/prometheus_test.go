package exporter

import (
	"context"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
)

func TestNewPrometheus(t *testing.T) {
	tests := []struct {
		name string
		opts PrometheusOptions
	}{
		{
			name: "default",
			opts: PrometheusOptions{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn := NewPrometheus(tt.opts)

			r, err := fn(context.TODO())
			assert.NoError(t, err)
			assert.NotNil(t, r)
		})
	}

	t.Run("RegisterWithSameRegistryTwice", func(t *testing.T) {
		reg := prometheus.NewRegistry()
		fn := NewPrometheus(PrometheusOptions{
			Registerer: reg,
		})

		_, err := fn(context.Background())
		assert.NoError(t, err)

		_, err = fn(context.Background())
		assert.Error(t, err)
	})
}
