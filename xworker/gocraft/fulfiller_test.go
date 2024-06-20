package gocraft

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/Bose/minisentinel"
	"github.com/alicebob/miniredis/v2"
	"github.com/gojek/work"
	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/suite"

	"github.com/gojekfarm/xtools/xworker"
)

type WorkerFulfillerSuite struct {
	suite.Suite

	UseRawEncodedPayload bool
	RedisPassword        string
}

func TestWorkerFulfillerSuite(t *testing.T) {
	suite.Run(t, new(WorkerFulfillerSuite))

	t.Run("UseRawEncodedPayloadisTrue", func(t *testing.T) {
		s := new(WorkerFulfillerSuite)
		s.UseRawEncodedPayload = true

		suite.Run(t, s)
	})

	t.Run("WithRedisPassword", func(t *testing.T) {
		s := new(WorkerFulfillerSuite)
		s.RedisPassword = "password"

		suite.Run(t, s)
	})
}

func (s *WorkerFulfillerSuite) TestNewWorkerAdapter() {
	r, err := miniredis.Run()
	s.NoError(err)

	gc := New(Options{
		Pool: &redis.Pool{Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", r.Addr())
		}},
		UseRawEncodedPayload: s.UseRawEncodedPayload,
	})
	s.NotNil(gc.WebUIServer())

	a, err := xworker.NewAdapter(xworker.AdapterOptions{Fulfiller: gc})
	s.NoError(err)

	errCount := 0

	done := make(chan struct{}, 1)

	s.NoError(a.RegisterHandlerWithOptions(
		"test-job",
		xworker.HandlerFunc(func(ctx context.Context, job *xworker.Job) error {
			if errCount == 0 {
				errCount++
				return errors.New("random_err")
			}
			var tp testPayload
			s.NoError(job.DecodePayload(&tp))

			s.Equal("John", tp.FieldA)
			s.Equal("Doe", tp.FieldB)

			done <- struct{}{}
			return nil
		}),
		xworker.RegisterOptions{
			MaxRetries:           3,
			RetryBackoffStrategy: xworker.ConstantRetryBackoff(time.Second),
		},
	))

	s.NoError(a.Start())
	defer func() {
		s.NoError(a.Stop())
	}()

	_, err = a.Enqueue(context.Background(), &xworker.Job{
		Name: "test-job",
		Payload: &testPayload{
			FieldA: "John",
			FieldB: "Doe",
		},
	})

	s.NoError(err)

	<-done
}

func (s *WorkerFulfillerSuite) TestNewWorkerAdapterWithoutCodec() {
	r, err := miniredis.Run()
	s.NoError(err)

	gc := New(Options{
		Pool: &redis.Pool{Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", r.Addr())
		}},
		DisableCodec:         true,
		UseRawEncodedPayload: s.UseRawEncodedPayload,
	})
	s.NotNil(gc.WebUIServer())

	a, err := xworker.NewAdapter(xworker.AdapterOptions{Fulfiller: gc})
	s.NoError(err)

	errCount := 0

	done := make(chan struct{}, 1)

	s.NoError(a.RegisterHandlerWithOptions(
		"test-job",
		xworker.HandlerFunc(func(ctx context.Context, job *xworker.Job) error {
			if errCount == 0 {
				errCount++
				return errors.New("random_err")
			}

			val := map[string]interface{}{}
			s.NoError(job.DecodePayload(&val))

			s.Equal("John", val["FieldA"])
			s.Equal("Doe", val["FieldB"])

			done <- struct{}{}
			return nil
		}),
		xworker.RegisterOptions{
			MaxRetries:           3,
			RetryBackoffStrategy: xworker.ConstantRetryBackoff(time.Second),
		},
	))

	s.NoError(a.Start())
	defer func() {
		s.NoError(a.Stop())
	}()

	_, err = a.Enqueue(context.Background(), &xworker.Job{
		Name: "test-job",
		Payload: map[string]interface{}{
			"FieldA": "John",
			"FieldB": "Doe",
		},
	})

	s.NoError(err)

	<-done
}

func (s *WorkerFulfillerSuite) TestNewWorkerAdapterWithoutCodecWithCustomDecode() {
	r, err := miniredis.Run()
	s.NoError(err)

	gc := New(Options{
		Pool: &redis.Pool{Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", r.Addr())
		}},
		DisableCodec:         true,
		UseRawEncodedPayload: s.UseRawEncodedPayload,
	})
	s.NotNil(gc.WebUIServer())

	a, err := xworker.NewAdapter(xworker.AdapterOptions{Fulfiller: gc})
	s.NoError(err)

	errCount := 0

	done := make(chan struct{}, 1)

	type testStruct struct {
		FieldA string        `json:"field_a"`
		FieldB int           `json:"field_b"`
		FieldC time.Duration `json:"field_c"`
	}

	s.NoError(a.RegisterHandlerWithOptions(
		"test-job",
		xworker.HandlerFunc(func(ctx context.Context, job *xworker.Job) error {
			if errCount == 0 {
				errCount++
				return errors.New("random_err")
			}

			val := &testStruct{}
			s.NoError(job.DecodePayload(val))

			s.Equal("John", val.FieldA)
			s.Equal(2, val.FieldB)
			s.Equal(time.Minute, val.FieldC)

			done <- struct{}{}
			return nil
		}),
		xworker.RegisterOptions{
			MaxRetries:           3,
			RetryBackoffStrategy: xworker.ConstantRetryBackoff(time.Second),
		},
	))

	s.NoError(a.Start())
	defer func() {
		s.NoError(a.Stop())
	}()

	_, err = a.Enqueue(context.Background(), &xworker.Job{
		Name: "test-job",
		Payload: map[string]interface{}{
			"field_a": "John",
			"field_b": 2,
			"field_c": time.Minute,
		},
	})

	s.NoError(err)

	<-done
}

func (s *WorkerFulfillerSuite) TestNewWorkerAdapterWithSentinel() {
	sntl := s.newTestSentinel()
	a, _ := s.newTestAdapterWithSentinel(sntl, false)

	done := make(chan struct{}, 1)

	s.NoError(a.RegisterHandlerWithOptions(
		"test-job-sentinel",
		xworker.HandlerFunc(func(ctx context.Context, job *xworker.Job) error {
			s.NoError(job.Raw.(*work.Job).ArgError())

			done <- struct{}{}
			return nil
		}),
		xworker.RegisterOptions{},
	))

	s.NoError(a.Start())
	defer func() {
		s.NoError(a.Stop())
	}()

	time.Sleep(100 * time.Millisecond)

	_, err := a.Enqueue(context.Background(), &xworker.Job{
		Name: "test-job-sentinel",
		Payload: map[string]interface{}{
			"arg1": "John",
			"arg2": "Doe",
		},
	})

	s.NoError(err)

	<-done
}

func (s *WorkerFulfillerSuite) TestSentinelFailOver() {
	sntl := s.newTestSentinel()
	a, f := s.newTestAdapterWithSentinel(sntl, true)
	mc := &testMetricCollector{}

	f.SetFailOverDurationCollector(mc.CollectFailOverDuration)

	s.NoError(a.Start())
	defer func() {
		s.NoError(a.Stop())
	}()

	time.Sleep(100 * time.Millisecond)

	startedAt := s.findStartTime(sntl)

	s.simulateFailOver(sntl)

	time.Sleep(11 * time.Second) // max jitter++
	s.Greater(s.findStartTime(sntl), startedAt)

	mc.mu.RLock()
	defer mc.mu.RUnlock()

	s.Len(mc.failOverDurations, 1)
}

func (s *WorkerFulfillerSuite) newTestAdapterWithSentinel(sntl *minisentinel.Sentinel, restartOnFailover bool) (xworker.Fulfiller, *Fulfiller) {
	s.T().Helper()

	f := NewWithSentinel(Options{
		Namespace:            "go-sentinel-worker-test",
		MaxConcurrency:       5,
		UseRawEncodedPayload: s.UseRawEncodedPayload,
	}, SentinelPoolOptions{
		Addresses:                   []string{sntl.Addr()},
		MaxDialAttempts:             10,
		MaxActiveConnections:        100,
		MaxIdleConnections:          10,
		IdleConnectionTimeout:       time.Second * 10,
		DialBackoffDelay:            time.Second,
		RestartWorkerPoolOnFailOver: restartOnFailover,
		RedisPassword:               s.RedisPassword,
	})

	a, err := xworker.NewAdapter(xworker.AdapterOptions{
		Fulfiller: f,
	})

	s.NoError(err)
	return a, f
}

func (s *WorkerFulfillerSuite) newTestSentinel() *minisentinel.Sentinel {
	s.T().Helper()

	a, err := miniredis.Run()
	s.NoError(err)
	s.T().Cleanup(a.Close)

	b, err := miniredis.Run()
	s.NoError(err)
	s.T().Cleanup(b.Close)
	a.RequireAuth(s.RedisPassword)
	b.RequireAuth(s.RedisPassword)

	sentinel := minisentinel.NewSentinel(a, minisentinel.WithReplica(b))
	err = sentinel.Start()
	s.NoError(err)
	s.T().Cleanup(sentinel.Close)

	return sentinel
}

func (s *WorkerFulfillerSuite) findStartTime(sntl *minisentinel.Sentinel) int64 {
	conn, err := redis.Dial("tcp", sntl.Master().Addr(), redis.DialPassword(s.RedisPassword))
	s.NoError(err)

	poolIDs, err := redis.Strings(conn.Do("SMEMBERS", "go-sentinel-worker-test:worker_pools"))
	s.NoError(err)
	if !s.Len(poolIDs, 1) {
		return 0
	}

	// get started at timestamp from heartbeat
	startedAt, err := redis.Int64(
		conn.Do("HGET", "go-sentinel-worker-test:worker_pools:"+poolIDs[0], "started_at"),
	)
	s.NoError(err)

	return startedAt
}

func (s *WorkerFulfillerSuite) simulateFailOver(sntl *minisentinel.Sentinel) {
	sntl.Lock()
	defer sntl.Unlock()

	master, slave := sntl.Master(), sntl.Replica()
	sntl.WithMaster(slave)

	// Kill existing connections open from adapter to this redis instance.
	master.Close()
	err := master.Start()
	s.NoError(err)

	sntl.SetReplica(master)
}

type testPayload struct {
	FieldA string `json:"field_a"`
	FieldB string `json:"field_b"`
}
