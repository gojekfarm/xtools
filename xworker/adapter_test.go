package xworker

import (
	"bytes"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/sethvargo/go-retry"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type testPayload struct {
	FieldOne string `json:"field_one"`
	FieldTwo int    `json:"field_two"`
}

func TestNewAdapter(t *testing.T) {
	mf := newMockFulfiller(t)
	mf.On("SetFailOverDurationCollector", mock.AnythingOfType("func(string, time.Duration)")).Return()

	bs, _ := retry.NewConstant(10 * time.Millisecond)
	a, err := NewAdapter(AdapterOptions{
		Fulfiller:       mf,
		MetricCollector: mf,
		RetryBackoff:    retry.WithMaxRetries(3, bs),
		OnRetryError:    func(err error) { t.Log(err) },
	})
	assert.NoError(t, err)

	done := make(chan struct{}, 1)

	mf.On("RegisterHandlerWithOptions",
		"test-job",
		mock.AnythingOfType("xworker.HandlerFunc"),
		RegisterOptions{},
	).Return(nil)

	assert.NoError(t, a.RegisterHandlerWithOptions("test-job", HandlerFunc(func(ctx context.Context, job *Job) error {
		var tp testPayload
		assert.NoError(t, job.DecodePayload(&tp))

		assert.Equal(t, "John Doe", tp.FieldOne)
		assert.Equal(t, 30, tp.FieldTwo)

		done <- struct{}{}
		return nil
	}), RegisterOptions{}))

	mf.On("Start").Return(nil)
	assert.NoError(t, a.Start())

	// Return 2 errors to check if retry works.
	mf.On("Enqueue", mock.Anything, mock.MatchedBy(func(j *Job) bool {
		return j.Name == "test-job"
	}), mock.Anything).Return(nil, errors.New("redis_err")).Twice()

	mf.On("Enqueue", mock.Anything, mock.MatchedBy(func(j *Job) bool {
		return j.Name == "test-job"
	}), mock.Anything).Return(NewEnqueueResult("mockFulfiller", "MockTask"), nil)

	er, err := a.Enqueue(context.Background(), &Job{
		Name: "test-job",
		Payload: testPayload{
			FieldOne: "John Doe",
			FieldTwo: 30,
		},
	})

	assert.IsType(t, "", er.Value())
	assert.Equal(t, "mockFulfiller", er.String())
	assert.NoError(t, err)

	<-done

	mf.On("Stop").Return(nil)
	assert.NoError(t, a.Stop())

	mf.AssertExpectations(t)
}

func TestAdapter_Run(t *testing.T) {
	mf := newMockFulfiller(t)
	a, err := NewAdapter(AdapterOptions{Fulfiller: mf})
	assert.NoError(t, err)

	ctx, cancel := context.WithCancel(context.TODO())
	errCh := make(chan error, 1)

	mf.On("Start").Return(errors.New("start error")).Once()
	assert.Error(t, a.Run(ctx))

	mf.On("Start").Return(nil).Once()
	go func() {
		errCh <- a.Run(ctx)
	}()

	mf.On("Stop").Return(nil).Once()

	cancel()

	assert.NoError(t, <-errCh)
}

func TestNewAdapterErrorWhenFulfillerDoesNotImplementWorkerUpdateFetcher(t *testing.T) {
	ao, err := NewAdapter(AdapterOptions{
		Fulfiller:       nil,
		MetricCollector: newMockFulfiller(t),
	})
	assert.Nil(t, ao)
	assert.Error(t, err)
}

func TestAdapter_Start_Error(t *testing.T) {
	mf := newMockFulfiller(t)
	mf.On("Start").Return(errors.New("redis_err"))
	a := &Adapter{fulfiller: mf}
	assert.Error(t, a.Start())
}

func TestAdapter_Stop_Error(t *testing.T) {
	mf := newMockFulfiller(t)
	mf.On("Stop").Return(errors.New("redis_err"))

	a := &Adapter{fulfiller: mf}
	a.metricsShutdown = func() {}

	assert.Error(t, a.Stop())
}

// newMockFulfiller can be used while testing where an implementation of Fulfiller is required.
func newMockFulfiller(t *testing.T) *mockFulfiller {
	m := &mockFulfiller{
		enqueueCh: make(chan *Job),
		executeCh: make(chan *Job),
		stopCh:    make(chan struct{}),
	}
	m.Test(t)

	return m
}

// mockFulfiller is intentionally not exported to avoid generating docs for this type in public API.
type mockFulfiller struct {
	mock.Mock

	enqueueCh chan *Job
	executeCh chan *Job

	stopCh chan struct{}
}

func (m *mockFulfiller) Enqueue(
	ctx context.Context,
	job *Job,
	opts ...Option,
) (*EnqueueResult, error) {
	args := m.Called(ctx, job, opts)

	er := args.Get(0)
	if er == nil {
		return nil, args.Error(1)
	}

	err := args.Error(1)
	if err == nil {
		m.enqueueCh <- job
	}

	return er.(*EnqueueResult), args.Error(1)
}

func (m *mockFulfiller) RegisterHandlerWithOptions(
	jobName string,
	jobHandler Handler,
	options RegisterOptions,
) error {
	go func() {
		for {
			select {
			case j := <-m.executeCh:
				if j.Name == jobName {
					var buf bytes.Buffer
					_, _ = buf.ReadFrom(j)
					_, _ = j.Write(buf.Bytes())
					_ = jobHandler.Handle(context.Background(), j)
				}
			case <-m.stopCh:
				return
			}
		}
	}()

	return m.Called(jobName, jobHandler, options).Error(0)
}

func (m *mockFulfiller) EnqueuePeriodically(cronSchedule string, job *Job, option ...Option) error {
	return m.Called(cronSchedule, job, option).Error(0)
}

func (m *mockFulfiller) Start() error {
	go func() {
		for {
			select {
			case j := <-m.enqueueCh:
				m.executeCh <- j
			case <-m.stopCh:
				return
			}
		}
	}()

	return m.Called().Error(0)
}

func (m *mockFulfiller) Stop() error {
	close(m.stopCh)

	return m.Called().Error(0)
}

func (m *mockFulfiller) FetchWorkerUpdate() (*WorkerUpdate, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}

	return args.Get(0).(*WorkerUpdate), args.Error(1)
}

func (m *mockFulfiller) CollectWorkerUpdate(wu WorkerUpdate) {
	m.Called(wu)
}

func (m *mockFulfiller) CollectFailOverDuration(ns string, d time.Duration) {
	m.Called(ns, d)
}

func (m *mockFulfiller) SetFailOverDurationCollector(fn func(string, time.Duration)) {
	m.Called(fn)
}
