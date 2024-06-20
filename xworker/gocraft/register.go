package gocraft

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"reflect"

	"github.com/gojek/work"

	"github.com/gojekfarm/xtools/xworker"
)

// RegisterHandlerWithOptions adds the HandlerFunc for a Job with given jobName and xworker.RegisterOptions.
func (f *Fulfiller) RegisterHandlerWithOptions(
	jobName string,
	handler xworker.Handler,
	opts xworker.RegisterOptions,
) error {
	f.pool.JobWithOptions(jobName, work.JobOptions{
		Priority:       uint(opts.Priority),
		MaxFails:       uint(opts.MaxRetries + 1),
		SkipDead:       opts.SkipArchive,
		MaxConcurrency: opts.MaxConcurrency,
		Backoff:        f.backoffCalculatorFunc(opts),
	}, f.jobHandler(handler))

	return nil
}

func (f *Fulfiller) jobHandler(jobHandler xworker.Handler) func(wj *work.Job) error {
	return func(wj *work.Job) error {
		j, err := f.parseJob(wj)
		if err != nil {
			return err
		}

		return jobHandler.Handle(context.Background(), j)
	}
}

func (f *Fulfiller) parseJob(wj *work.Job) (*xworker.Job, error) {
	j := &xworker.Job{Name: wj.Name, Raw: wj}

	if err := f.decodeJobPayload(wj, j); err != nil {
		return nil, err
	}

	return j, nil
}

func (f *Fulfiller) decodeJobPayload(wj *work.Job, w io.Writer) error {
	if !f.hasJobBlob(wj) {
		return json.NewEncoder(w).Encode(wj.Args)
	}

	b, err := f.readJobBlob(wj)
	if err != nil {
		return err
	}

	_, err = w.Write(b.payload())

	return err
}

func (f *Fulfiller) backoffCalculatorFunc(opts xworker.RegisterOptions) work.BackoffCalculator {
	if opts.RetryBackoffStrategy == nil {
		return nil
	}

	return func(wj *work.Job) int64 {
		j, err := f.parseJob(wj)
		if err != nil {
			return 0
		}

		return int64(opts.RetryBackoffStrategy.RetryBackoff(int(wj.Fails-1), errors.New(wj.LastErr), j).Seconds())
	}
}

func (f *Fulfiller) readJobBlob(j *work.Job) (*jobBlob, error) {
	jb := &jobBlob{}

	if err := json.Unmarshal(reflect.ValueOf(j).Elem().FieldByName("rawJSON").Bytes(), jb); err != nil {
		return nil, err
	}

	return jb, nil
}

func (f *Fulfiller) hasJobBlob(j *work.Job) bool {
	_, hasPayloadByteKey := j.Args[payloadByteKey]
	_, hasPayloadRawKey := j.Args[payloadRawKey]

	return (hasPayloadByteKey || hasPayloadRawKey) &&
		len(j.Args) == 1 // Guard for cases where map already has `payload` key, application developer to handle
}
