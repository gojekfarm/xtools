package prometheus

import (
	"github.com/prometheus/client_golang/prometheus"
)

// RegisterConsumerMetrics registers all prometheus.Collector(s) for consumer with given prometheus.Registerer.
func RegisterConsumerMetrics(r prometheus.Registerer) error {
	for _, c := range allConsumerCollectors {
		if err := r.Register(c); err != nil {
			return err
		}
	}

	return nil
}

// RegisterProducerMetrics registers all prometheus.Collector(s) for producer with given prometheus.Registerer.
func RegisterProducerMetrics(r prometheus.Registerer) error {
	for _, p := range allProducerCollectors {
		if err := r.Register(p); err != nil {
			return err
		}
	}

	return nil
}
