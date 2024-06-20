package gocraft

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/assert"

	"github.com/gojekfarm/xtools/xworker"
)

func TestWorkerAdapter_collectWorkerUpdatesPeriodically(t *testing.T) {
	r, err := miniredis.Run()
	assert.NoError(t, err)

	mc := &testMetricCollector{}
	a, err := xworker.NewAdapter(xworker.AdapterOptions{
		Fulfiller: New(Options{
			Pool: &redis.Pool{Dial: func() (redis.Conn, error) {
				return redis.Dial("tcp", r.Addr())
			}},
		}),
		MetricCollector: mc,
	})
	assert.NoError(t, err)

	done := make(chan struct{}, 1)

	assert.NoError(t, a.RegisterHandlerWithOptions("test-job",
		xworker.HandlerFunc(func(ctx context.Context, job *xworker.Job) error {
			var tp testPayload
			assert.NoError(t, job.DecodePayload(&tp))

			assert.Equal(t, "John", tp.FieldA)
			assert.Equal(t, "Doe", tp.FieldB)

			done <- struct{}{}
			return nil
		}), xworker.RegisterOptions{}))

	assert.NoError(t, a.Start())
	defer func() {
		assert.NoError(t, a.Stop())
	}()

	_, err = a.Enqueue(context.Background(), &xworker.Job{
		Name: "test-job",
		Payload: &testPayload{
			FieldA: "John",
			FieldB: "Doe",
		},
	})

	assert.NoError(t, err)

	<-done

	assert.Eventually(t, func() bool {
		mc.mu.RLock()
		defer mc.mu.RUnlock()
		return len(mc.workerUpdates) > 1
	}, 3*time.Second, 100*time.Millisecond, "CollectWorkerUpdate no called")
}

type testMetricCollector struct {
	workerUpdates     []xworker.WorkerUpdate
	failOverDurations []time.Duration

	mu sync.RWMutex
}

func (t *testMetricCollector) CollectWorkerUpdate(update xworker.WorkerUpdate) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.workerUpdates = append(t.workerUpdates, update)
}

func (t *testMetricCollector) CollectFailOverDuration(_ string, duration time.Duration) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.failOverDurations = append(t.failOverDurations, duration)
}
