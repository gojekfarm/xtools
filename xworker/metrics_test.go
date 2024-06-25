package xworker

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestAdapter_collectWorkerUpdatesPeriodically(t *testing.T) {
	twu := &WorkerUpdate{
		Namespace:  "test-namespace",
		DeadCount:  10,
		RetryCount: 5,
		QueueMetrics: []QueueMetric{
			{Name: "test-queue", Count: 50, MaxConcurrency: 25, Latency: time.Second, LockCount: 5},
		},
	}

	tests := []struct {
		name                  string
		mockFulfiller         func(*testing.T) *mockFulfiller
		metricCollectInterval time.Duration
		times                 int
	}{
		{
			name: "DefaultMetricInterval",
			mockFulfiller: func(t *testing.T) *mockFulfiller {
				mf := newMockFulfiller(t)
				mf.On("FetchWorkerUpdate").Return(twu, nil).Once()
				mf.On("CollectWorkerUpdate", *twu).Return().Once()
				return mf
			},
		},
		{
			name:                  "CustomMetricInterval",
			metricCollectInterval: 500 * time.Millisecond,
			mockFulfiller: func(t *testing.T) *mockFulfiller {
				mf := newMockFulfiller(t)
				mf.On("FetchWorkerUpdate").Return(twu, nil).Once()
				mf.On("CollectWorkerUpdate", *twu).Return().Once()
				return mf
			},
		},
		{
			name:                  "MultipleCollectCalls",
			metricCollectInterval: 500 * time.Millisecond,
			mockFulfiller: func(t *testing.T) *mockFulfiller {
				mf := newMockFulfiller(t)
				mf.On("FetchWorkerUpdate").Return(twu, nil).Times(10)
				mf.On("CollectWorkerUpdate", *twu).Return().Times(10)
				return mf
			},
			times: 9,
		},
		{
			name: "ErrorInFetchWorkerUpdate",
			mockFulfiller: func(t *testing.T) *mockFulfiller {
				mf := newMockFulfiller(t)
				mf.On("FetchWorkerUpdate").
					Return(nil, errors.New("redis err")).Once()
				return mf
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mf := tt.mockFulfiller(t)
			a := &Adapter{
				workerUpdateFetcher:   mf,
				metricCollector:       mf,
				metricCollectInterval: tt.metricCollectInterval,
			}

			a.metricsCtx, a.metricsShutdown = context.WithCancel(context.Background())

			go a.collectWorkerUpdatesPeriodically()

			for i := 0; i < tt.times+1; i++ {
				if tt.metricCollectInterval > 0 {
					time.Sleep(tt.metricCollectInterval)
				} else {
					// Since default interval is 1s
					time.Sleep(time.Second)
				}
			}

			time.Sleep(100 * time.Millisecond)

			a.metricsShutdown()

			mf.AssertExpectations(t)
		})
	}
}
