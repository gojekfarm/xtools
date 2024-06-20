package gocraft

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gojek/work"
	"github.com/gojekfarm/xtools/xworker"
	"github.com/gomodule/redigo/redis"
	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type mapAlias map[string]interface{}

type testcase struct {
	name         string
	job          *xworker.Job
	options      []xworker.Option
	enqueuerMock func(*mock.Mock, *xworker.Job)
	wantErr      bool
}

func TestWorkerAdapter_Enqueue(t *testing.T) {
	payloadByteKeyMatcher := mock.MatchedBy(func(args map[string]interface{}) bool {
		_, ok := args[payloadByteKey]
		return ok
	})

	payloadRawKeyMatcher := mock.MatchedBy(func(args map[string]interface{}) bool {
		_, ok := args[payloadRawKey]
		return ok
	})

	codecTests := enqueueTestcases("Codec", payloadByteKeyMatcher)
	nonCodecTests := enqueueTestcases("NoCodec", mock.MatchedBy(func(args map[string]interface{}) bool {
		_, ok1 := args["arg1"]
		_, ok2 := args["arg2"]
		return ok1 && ok2
	}))
	rawCodecTests := enqueueTestcases("RawCodec", payloadRawKeyMatcher)

	for _, tt := range codecTests {
		t.Run(tt.name, func(t *testing.T) {
			runEnqueueTest(t, tt, true, false)
		})
	}

	for _, tt := range nonCodecTests {
		t.Run(tt.name, func(t *testing.T) {
			runEnqueueTest(t, tt, false, false)
		})
	}

	for _, tt := range rawCodecTests {
		t.Run(tt.name, func(t *testing.T) {
			runEnqueueTest(t, tt, true, true)
		})
	}
}

func runEnqueueTest(t *testing.T, tt testcase, useCodec bool, useRawEncodedPayload bool) {
	em := newMockEnqueuer(t)

	a, err := xworker.NewAdapter(xworker.AdapterOptions{
		Fulfiller: &Fulfiller{
			enqueuer:             em,
			disableCodec:         !useCodec,
			useRawEncodedPayload: useRawEncodedPayload,
		},
	})
	assert.NoError(t, err)

	if tt.enqueuerMock != nil {
		tt.enqueuerMock(&em.Mock, tt.job)
	}

	_, err = a.Enqueue(context.Background(), tt.job, tt.options...)
	if tt.wantErr {
		assert.Error(t, err)
	} else {
		assert.NoError(t, err)
	}

	em.AssertExpectations(t)
}

func enqueueTestcases(prefix string, argMatcher interface{}) []testcase {
	tests := []testcase{
		{
			name: prefix + "Enqueue",
			job: &xworker.Job{
				Name: "test-job",
				Payload: map[string]interface{}{
					"arg1": "arg1",
					"arg2": "arg2",
				},
			},
			enqueuerMock: func(m *mock.Mock, job *xworker.Job) {
				m.On("Enqueue", job.Name, argMatcher).Return(&work.Job{}, nil)
			},
		},
		{
			name: prefix + "EnqueueUnique",
			job: &xworker.Job{
				Name: "test-job",
				Payload: map[string]interface{}{
					"arg1": "arg1",
					"arg2": "arg2",
				},
			},
			options: []xworker.Option{xworker.Unique},
			enqueuerMock: func(m *mock.Mock, job *xworker.Job) {
				m.On("EnqueueUnique", job.Name, argMatcher).Return(&work.Job{}, nil)
			},
		},
		{
			name: prefix + "EnqueueError",
			job: &xworker.Job{
				Name: "test-job",
				Payload: map[string]interface{}{
					"arg1": "arg1",
					"arg2": "arg2",
				},
			},
			enqueuerMock: func(m *mock.Mock, job *xworker.Job) {
				m.On("Enqueue", job.Name, argMatcher).
					Return(nil, errors.New("redis_err"))
			},
			wantErr: true,
		},
		{
			name: prefix + "EnqueueIn",
			job: &xworker.Job{
				Name: "test-job",
				Payload: map[string]interface{}{
					"arg1": "arg1",
					"arg2": "arg2",
				},
			},
			options: []xworker.Option{xworker.In(time.Second)},
			enqueuerMock: func(m *mock.Mock, job *xworker.Job) {
				m.On("EnqueueIn", job.Name, int64(1), argMatcher).
					Return(&work.ScheduledJob{Job: &work.Job{
						Name:       "test-job",
						ID:         "test-job",
						EnqueuedAt: int64(time.Now().Second()),
					}, RunAt: int64(time.Now().Add(time.Second).Second())}, nil)
			},
		},
		{
			name: prefix + "EnqueueUniqueIn",
			job: &xworker.Job{
				Name: "test-job",
				Payload: map[string]interface{}{
					"arg1": "arg1",
					"arg2": "arg2",
				},
			},
			options: []xworker.Option{xworker.In(time.Second), xworker.Unique},
			enqueuerMock: func(m *mock.Mock, job *xworker.Job) {
				m.On("EnqueueUniqueIn", job.Name, int64(1), argMatcher).
					Return(&work.ScheduledJob{Job: &work.Job{
						Name:       "test-job",
						ID:         "test-job",
						EnqueuedAt: int64(time.Now().Second()),
					}, RunAt: int64(time.Now().Add(time.Second).Second())}, nil)
			},
		},
		{
			name: prefix + "EnqueueInError",
			job: &xworker.Job{
				Name: "test-job",
				Payload: map[string]interface{}{
					"arg1": "arg1",
					"arg2": "arg2",
				},
			},
			options: []xworker.Option{xworker.In(time.Second)},
			enqueuerMock: func(m *mock.Mock, job *xworker.Job) {
				m.On("EnqueueIn", job.Name, int64(1), argMatcher).
					Return(nil, errors.New("redis_err"))
			},
			wantErr: true,
		},
		{
			name: prefix + "EnqueueAt",
			job: &xworker.Job{
				Name: "test-job",
				Payload: map[string]interface{}{
					"arg1": "arg1",
					"arg2": "arg2",
				},
			},
			options: []xworker.Option{xworker.At(time.Now().Add(2 * time.Second))},
			enqueuerMock: func(m *mock.Mock, job *xworker.Job) {
				m.On("EnqueueIn", job.Name, int64(1), argMatcher).
					Return(&work.ScheduledJob{Job: &work.Job{
						Name:       "test-job",
						ID:         "test-job",
						EnqueuedAt: int64(time.Now().Second()),
					}, RunAt: int64(time.Now().Add(2 * time.Second).Second())}, nil)
			},
		},
		{
			name: prefix + "EnqueueUniqueAt",
			job: &xworker.Job{
				Name: "test-job",
				Payload: map[string]interface{}{
					"arg1": "arg1",
					"arg2": "arg2",
				},
			},
			options: []xworker.Option{xworker.At(time.Now().Add(2 * time.Second)), xworker.Unique},
			enqueuerMock: func(m *mock.Mock, job *xworker.Job) {
				m.On("EnqueueUniqueIn", job.Name, int64(1), argMatcher).
					Return(&work.ScheduledJob{Job: &work.Job{
						Name:       "test-job",
						ID:         "test-job",
						EnqueuedAt: int64(time.Now().Second()),
					}, RunAt: int64(time.Now().Add(2 * time.Second).Second())}, nil)
			},
		},
		{
			name: prefix + "EnqueueAtError",
			job: &xworker.Job{
				Name: "test-job",
				Payload: map[string]interface{}{
					"arg1": "arg1",
					"arg2": "arg2",
				},
			},
			options: []xworker.Option{xworker.At(time.Now().Add(2 * time.Second))},
			enqueuerMock: func(m *mock.Mock, job *xworker.Job) {
				m.On("EnqueueIn", job.Name, int64(1), argMatcher).
					Return(nil, errors.New("redis_err"))
			},
			wantErr: true,
		},
		{
			name: prefix + "EnqueueUniqueDuplicate",
			job: &xworker.Job{
				Name: "test-job",
				Payload: map[string]interface{}{
					"arg1": "arg1",
					"arg2": "arg2",
				},
			},
			options: []xworker.Option{xworker.Unique},
			enqueuerMock: func(m *mock.Mock, job *xworker.Job) {
				m.On("EnqueueUnique", job.Name, argMatcher).
					// Job and error can be nil when the job is duplicate
					Return(nil, nil)
			},
		},
		{
			name: prefix + "EnqueueUniqueInDuplicate",
			job: &xworker.Job{
				Name: "test-job",
				Payload: map[string]interface{}{
					"arg1": "arg1",
					"arg2": "arg2",
				},
			},
			options: []xworker.Option{xworker.Unique, xworker.In(time.Minute)},
			enqueuerMock: func(m *mock.Mock, job *xworker.Job) {
				m.On("EnqueueUniqueIn", job.Name, int64(60), argMatcher).
					// Job and error can be nil when the job is duplicate
					Return(nil, nil)
			},
		},
	}

	return tests
}

func TestWorkerAdapter_EnqueuePeriodically(t *testing.T) {
	r, err := miniredis.Run()
	assert.NoError(t, err)

	wa := New(Options{
		Pool: &redis.Pool{Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", r.Addr())
		}},
		Logger: log.Logger,
	})

	assert.NoError(t, wa.Start())
	defer func() {
		assert.NoError(t, wa.Stop())
	}()

	assert.NoError(t, wa.EnqueuePeriodically("*/30 * * * *", &xworker.Job{Name: "test-job"}))
}

func TestFulfiller_enqueueWithoutCodec_nilPayload(t *testing.T) {
	me := newMockEnqueuer(t)
	me.On("Enqueue", "test-job", map[string]interface{}(nil)).Return(&work.Job{}, nil)

	f := &Fulfiller{enqueuer: me}

	_, err := f.enqueueWithoutCodec(&xworker.Job{Name: "test-job", Payload: nil}, nil)
	assert.NoError(t, err)

	me.AssertExpectations(t)
}

func TestFulfiller_enqueueWithoutCodec_mapAliasedPayload(t *testing.T) {
	me := newMockEnqueuer(t)
	me.On("Enqueue", "test-job", map[string]interface{}{"Key": "Val"}).Return(&work.Job{}, nil)

	f := &Fulfiller{enqueuer: me}

	_, err := f.enqueueWithoutCodec(&xworker.Job{Name: "test-job", Payload: mapAlias{"Key": "Val"}}, nil)
	assert.NoError(t, err)

	me.AssertExpectations(t)
}

func TestFulfiller_enqueueWithoutCodec_mapAliasedPayloadPointer(t *testing.T) {
	me := newMockEnqueuer(t)
	me.On("Enqueue", "test-job", map[string]interface{}{"Key": "Val"}).Return(&work.Job{}, nil)

	f := &Fulfiller{enqueuer: me}

	_, err := f.enqueueWithoutCodec(&xworker.Job{Name: "test-job", Payload: &mapAlias{"Key": "Val"}}, nil)
	assert.NoError(t, err)

	me.AssertExpectations(t)
}

func TestFulfiller_enqueueWithoutCodec_error(t *testing.T) {
	f := &Fulfiller{}

	_, err := f.enqueueWithoutCodec(&xworker.Job{Payload: testPayload{}}, nil)
	assert.EqualError(t, err, "payload gocraft.testPayload must be a type alias to map[string]interface{}")
}

type mockEnqueuer struct {
	mock.Mock
}

func newMockEnqueuer(t *testing.T) *mockEnqueuer {
	m := &mockEnqueuer{}
	m.Test(t)
	return m
}

func (m *mockEnqueuer) Enqueue(jobName string, args map[string]interface{}) (*work.Job, error) {
	arg := m.Called(jobName, args)
	if arg.Get(0) == nil {
		return nil, arg.Error(1)
	}
	return arg.Get(0).(*work.Job), arg.Error(1)
}

func (m *mockEnqueuer) EnqueueUnique(jobName string, args map[string]interface{}) (*work.Job, error) {
	arg := m.Called(jobName, args)
	if arg.Get(0) == nil {
		return nil, arg.Error(1)
	}
	return arg.Get(0).(*work.Job), arg.Error(1)
}

func (m *mockEnqueuer) EnqueueIn(jobName string, secondsFromNow int64, args map[string]interface{}) (*work.ScheduledJob, error) {
	arg := m.Called(jobName, secondsFromNow, args)
	if arg.Get(0) == nil {
		return nil, arg.Error(1)
	}
	return arg.Get(0).(*work.ScheduledJob), arg.Error(1)
}

func (m *mockEnqueuer) EnqueueUniqueIn(jobName string, secondsFromNow int64, args map[string]interface{}) (*work.ScheduledJob, error) {
	arg := m.Called(jobName, secondsFromNow, args)
	if arg.Get(0) == nil {
		return nil, arg.Error(1)
	}
	return arg.Get(0).(*work.ScheduledJob), arg.Error(1)
}
