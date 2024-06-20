package xworker

import "time"

// QueueMetric is an alias to work.Queue metrics.
type QueueMetric struct {
	// Name is the name of the queue.
	Name string `json:"name"`
	// Count is the size of the queue.
	Count int `json:"count"`
	// MaxConcurrency is the maximum concurrency that the queue can achieve.
	MaxConcurrency int `json:"max_concurrency"`
	// Latency is the measurement of how long ago the next job to be processed was enqueued.
	Latency time.Duration `json:"latency"`
	// LockCount is the queue lock count or current concurrency.
	LockCount int `json:"lock_count"`
}

// WorkerUpdate contains information about worker, usually sent in a periodic duration.
type WorkerUpdate struct {
	Namespace    string
	DeadCount    int
	RetryCount   int
	QueueMetrics []QueueMetric
}

// FailOverDurationCollectorSetter defines the function that can be exposed to
// relay the failOver duration to metrics collector.
type FailOverDurationCollectorSetter interface {
	SetFailOverDurationCollector(func(namespace string, duration time.Duration))
}

// WorkerUpdateFetcher is used to fetch WorkerUpdate from the worker implementation.
type WorkerUpdateFetcher interface {
	FetchWorkerUpdate() (*WorkerUpdate, error)
}

// MetricCollector helps in gathering WorkerUpdate metrics.
type MetricCollector interface {
	// CollectWorkerUpdate is called every AdapterOptions.MetricsCollectInterval duration,
	// this method should handle any read/write in a concurrency safe manner.
	CollectWorkerUpdate(WorkerUpdate)

	// CollectFailOverDuration is called on every sentinel failOver,
	// this method should handle any read/write in a concurrency safe manner.
	CollectFailOverDuration(namespace string, duration time.Duration)
}

func (a *Adapter) collectWorkerUpdatesPeriodically() {
	duration := time.Second
	if a.metricCollectInterval > 0 {
		duration = a.metricCollectInterval
	}

	ticker := time.NewTicker(duration)
	defer ticker.Stop()

	for {
		select {
		case <-a.metricsCtx.Done():
			return
		case <-ticker.C:
			wu, err := a.workerUpdateFetcher.FetchWorkerUpdate()
			if err != nil {
				a.logger.Warn().Err(err)

				continue
			}

			a.metricCollector.CollectWorkerUpdate(*wu)
		}
	}
}
