package exporter

import (
	pc "github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"

	"github.com/gojekfarm/xtools/xtel"
)

// PrometheusOptions handles the configuration of Prometheus exporter.
type PrometheusOptions struct {
	Registerer pc.Registerer
}

// NewPrometheus returns a new prometheus metric.Reader. It configs the new exporter with the given PrometheusOptions.
func NewPrometheus(opts PrometheusOptions) xtel.MetricReaderFunc {
	return func() (metric.Reader, error) {
		reg := opts.Registerer
		if reg == nil {
			reg = pc.DefaultRegisterer
		}

		return prometheus.New(prometheus.WithRegisterer(reg))
	}
}
