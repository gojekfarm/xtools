package exporter

import (
	"context"
	"sync"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type uptimeCollector struct {
	uptime      *prometheus.Desc
	uptimeValue int64
	mu          sync.RWMutex
}

func newUptimeCollector(ctx context.Context) prometheus.Collector {
	c := &uptimeCollector{
		uptime: prometheus.NewDesc(
			"system_wallclock",
			"The number of seconds passed since the node was started.",
			nil, nil),
		uptimeValue: 0,
	}

	go c.run(ctx)

	return c
}

func (c *uptimeCollector) run(ctx context.Context) {
	ticker := time.NewTicker(time.Second)

	for {
		select {
		case <-ctx.Done():
			ticker.Stop()

			return
		case <-ticker.C:
			c.incValue()
		}
	}
}

func (c *uptimeCollector) incValue() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.uptimeValue++
}

func (c *uptimeCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.uptime
}

func (c *uptimeCollector) Collect(ch chan<- prometheus.Metric) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	ch <- prometheus.MustNewConstMetric(c.uptime, prometheus.CounterValue, float64(c.uptimeValue))
}
