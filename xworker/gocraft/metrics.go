package gocraft

import (
	"time"

	"github.com/gojekfarm/xtools/xworker"
)

// SetFailOverDurationCollector sets the given function to be called when a sentinel failOver occurs.
func (f *Fulfiller) SetFailOverDurationCollector(fn func(namespace string, duration time.Duration)) {
	f.collectFailOverFunc = fn
}

// FetchWorkerUpdate returns xworker.WorkerUpdate for all work.Queue(s).
func (f *Fulfiller) FetchWorkerUpdate() (*xworker.WorkerUpdate, error) {
	_, dc, err := f.client.DeadJobs(1)
	if err != nil {
		return nil, err
	}

	_, rc, err := f.client.RetryJobs(1)
	if err != nil {
		return nil, err
	}

	queues, err := f.client.Queues()
	if err != nil {
		return nil, err
	}

	qm := make([]xworker.QueueMetric, 0, len(queues))
	for _, q := range queues {
		qm = append(qm, xworker.QueueMetric{
			Name:           q.JobName,
			Count:          int(q.Count),
			Latency:        time.Duration(q.Latency) * time.Second,
			MaxConcurrency: int(q.MaxConcurrency),
			LockCount:      int(q.LockCount),
		})
	}

	return &xworker.WorkerUpdate{
		Namespace:    f.namespace,
		DeadCount:    int(dc),
		RetryCount:   int(rc),
		QueueMetrics: qm,
	}, nil
}
