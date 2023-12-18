package exporter

import (
	"context"
	"os"
	"os/signal"

	pc "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"

	"github.com/gojekfarm/xtools/xtel"
)

// PrometheusOptions handles the configuration of Prometheus exporter.
type PrometheusOptions struct {
	Registerer               pc.Registerer
	DisableDefaultCollectors bool
}

// NewPrometheus returns a new prometheus metric.Reader. It configs the new exporter with the given PrometheusOptions.
func NewPrometheus(opts PrometheusOptions) xtel.MetricReaderFunc {
	return func(_ context.Context) (metric.Reader, error) {
		reg := opts.Registerer
		if reg == nil {
			reg = pc.DefaultRegisterer
		}

		if !opts.DisableDefaultCollectors {
			toRegister := []pc.Collector{uptimeCollectorInstance}
			if opts.Registerer != nil {
				toRegister = defaultCollectors
			}

			if err := registerCollectors(reg, toRegister...); err != nil {
				return nil, err
			}
		}

		return prometheus.New(prometheus.WithRegisterer(reg))
	}
}

func registerCollectors(reg pc.Registerer, collectors ...pc.Collector) error {
	for _, c := range collectors {
		if err := reg.Register(c); err != nil {
			return err
		}
	}

	return nil
}

var uptimeCollectorInstance pc.Collector
var defaultCollectors []pc.Collector

func init() {
	ctx, _ := signal.NotifyContext(context.Background(), os.Interrupt, os.Kill)
	uptimeCollectorInstance = newUptimeCollector(ctx)

	defaultCollectors = []pc.Collector{
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
		collectors.NewGoCollector(),
		uptimeCollectorInstance,
	}
}
