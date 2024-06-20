package gocraft

import (
	"context"
	"encoding/base64"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gomodule/redigo/redis"
	"github.com/stretchr/testify/suite"

	"github.com/gojekfarm/xtools/xworker"
)

type customJob struct {
	Bytes       []byte `json:"bytes"`
	StringBytes string `json:"string_bytes"`
	IntArray    []int  `json:"int_array"`
}

type MigrationSuite struct {
	suite.Suite

	enqueuerUseRawEncodedPayload   bool
	registererUseRawEncodedPayload bool

	mr    *miniredis.Miniredis
	e     xworker.Enqueuer
	r     xworker.Registerer
	start func() error
	stop  func() error
}

func TestMigrationSuite(t *testing.T) {
	tcs := []struct {
		name                           string
		enqueuerUseRawEncodedPayload   bool
		registererUseRawEncodedPayload bool
	}{
		{
			name:                           "UseRawEncodedPayload is false",
			enqueuerUseRawEncodedPayload:   false,
			registererUseRawEncodedPayload: false,
		}, {
			name:                           "UseRawEncodedPayload is true",
			enqueuerUseRawEncodedPayload:   true,
			registererUseRawEncodedPayload: true,
		}, {
			name:                           "UseRawEncodedPayload is true during enqueue and false during dequeue",
			enqueuerUseRawEncodedPayload:   true,
			registererUseRawEncodedPayload: false,
		}, {
			name:                           "UseRawEncodedPayload is false during enqueue and true during dequeue",
			enqueuerUseRawEncodedPayload:   false,
			registererUseRawEncodedPayload: true,
		},
	}

	for _, tc := range tcs {
		t.Run(tc.name, func(t *testing.T) {
			s := new(MigrationSuite)
			s.enqueuerUseRawEncodedPayload = tc.enqueuerUseRawEncodedPayload
			s.registererUseRawEncodedPayload = tc.registererUseRawEncodedPayload

			suite.Run(t, s)
		})
	}
}

type migrationTestcase struct {
	name    string
	payload interface{}
	decode  func(*MigrationSuite, *xworker.Job) error
}

func (s *MigrationSuite) TestFlowWithCodec() {
	const jobName = "encode-on-enqueue-and-decode-on-handle"

	testcases := append(s.commonArgsBasedTestcases(), s.customJobTestcase(false))

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			s.setupTestBed(true, true)

			s.runTest(jobName, tc)

			s.tearDownTestBed()
		})
	}
}

func (s *MigrationSuite) TestFlowWithoutCodec() {
	const jobName = "no-encode-on-enqueue-and-no-decode-on-handle"

	testcases := append(s.commonArgsBasedTestcases(), s.customJobTestcase(true))

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			s.setupTestBed(false, false)

			s.runTest(jobName, tc)

			s.tearDownTestBed()
		})
	}
}

func (s *MigrationSuite) TestHandleWithCodecAndEnqueueWithoutCodec() {
	const jobName = "no-encode-on-enqueue-and-decode-on-handle"

	testcases := append(s.commonArgsBasedTestcases(), s.customJobTestcase(true))

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			s.setupTestBed(false, true)

			s.runTest(jobName, tc)

			s.tearDownTestBed()
		})
	}
}

func (s *MigrationSuite) TestHandleWithoutCodecAndEnqueueWithCodec() {
	const jobName = "encode-on-enqueue-and-no-decode-on-handle"

	testcases := append(s.commonArgsBasedTestcases(), s.customJobTestcase(false))

	for _, tc := range testcases {
		s.Run(tc.name, func() {
			s.setupTestBed(true, false)

			s.runTest(jobName, tc)

			s.tearDownTestBed()
		})
	}
}

func (s *MigrationSuite) runTest(jobName string, tc migrationTestcase) {
	done := make(chan struct{}, 1)

	s.NoError(s.r.RegisterHandlerWithOptions(jobName,
		xworker.HandlerFunc(func(ctx context.Context, job *xworker.Job) error {
			defer func() {
				done <- struct{}{}
			}()

			if err := tc.decode(s, job); err != nil {
				return err
			}

			return nil
		}),
		xworker.RegisterOptions{}))

	s.NoError(s.start())

	_, err := s.e.Enqueue(context.Background(), &xworker.Job{
		Name:    jobName,
		Payload: tc.payload,
	})
	s.NoError(err)

	select {
	case <-time.After(5 * time.Second):
		s.Fail("Test timed out", tc.name)
	case <-done:
	}
}

func (s *MigrationSuite) customJobTestcase(mapBased bool) migrationTestcase {
	var p interface{}
	if mapBased {
		p = map[string]interface{}{
			"bytes":        []byte("random"),
			"string_bytes": "random",
			"int_array":    []int{1, 2, 3},
		}
	} else {
		p = &customJob{
			Bytes:       []byte("random"),
			StringBytes: "random",
			IntArray:    []int{1, 2, 3},
		}
	}

	return migrationTestcase{
		name:    "custom-job",
		payload: p,
		decode: func(s *MigrationSuite, job *xworker.Job) error {
			cj := new(customJob)
			err := job.DecodePayload(cj)
			s.NoError(err)
			s.Equal([]byte("random"), cj.Bytes)
			s.Equal("random", cj.StringBytes)
			s.EqualValues([]int{1, 2, 3}, cj.IntArray)
			return err
		},
	}
}

func (s *MigrationSuite) commonArgsBasedTestcases() []migrationTestcase {
	return []migrationTestcase{
		{
			name: "args-job",
			payload: map[string]interface{}{
				"key_1": "value_1",
				"key_2": 2,
				"key_3": []byte("random"),
				"key_4": []int{1, 2, 3},
			},
			decode: func(s *MigrationSuite, job *xworker.Job) error {
				aj := make(map[string]interface{})
				err := job.DecodePayload(&aj)
				s.NoError(err)
				s.Equal("value_1", aj["key_1"])
				s.EqualValues(2, aj["key_2"])
				b, err := base64.StdEncoding.DecodeString(aj["key_3"].(string))
				s.NoError(err)
				s.Equal([]byte("random"), b)
				for i, v := range aj["key_4"].([]interface{}) {
					s.EqualValues(i+1, v)
				}
				return err
			},
		},
		{
			name: "args-job-with-payload-key",
			payload: map[string]interface{}{
				"key_1":   "value_1",
				"key_2":   2,
				"key_3":   []byte("random"),
				"key_4":   []int{1, 2, 3},
				"payload": []byte("random_payload"),
			},
			decode: func(s *MigrationSuite, job *xworker.Job) error {
				aj := make(map[string]interface{})
				err := job.DecodePayload(&aj)
				s.NoError(err)
				s.Equal("value_1", aj["key_1"])
				s.EqualValues(2, aj["key_2"])
				b, err := base64.StdEncoding.DecodeString(aj["key_3"].(string))
				s.NoError(err)
				s.Equal([]byte("random"), b)
				p, err := base64.StdEncoding.DecodeString(aj["payload"].(string))
				s.NoError(err)
				s.Equal([]byte("random_payload"), p)
				for i, v := range aj["key_4"].([]interface{}) {
					s.EqualValues(i+1, v)
				}
				return err
			},
		},
	}
}

func (s *MigrationSuite) setupTestBed(encodeOnEnqueue, decodeOnHandle bool) {
	mr, err := miniredis.Run()
	s.NoError(err)
	s.mr = mr
	p := &redis.Pool{Dial: func() (redis.Conn, error) {
		return redis.Dial("tcp", mr.Addr())
	}}

	e, err := xworker.NewAdapter(xworker.AdapterOptions{Fulfiller: New(Options{
		Pool:                 p,
		DisableCodec:         !encodeOnEnqueue,
		UseRawEncodedPayload: s.enqueuerUseRawEncodedPayload,
	})})
	s.NoError(err)
	s.e = e

	r, err := xworker.NewAdapter(xworker.AdapterOptions{Fulfiller: New(Options{
		Pool:                 p,
		DisableCodec:         !decodeOnHandle,
		UseRawEncodedPayload: s.registererUseRawEncodedPayload,
	})})
	s.NoError(err)
	s.r = r

	s.start = r.Start
	s.stop = r.Stop
}

func (s *MigrationSuite) tearDownTestBed() {
	s.NoError(s.stop())
	s.mr.Close()
}
