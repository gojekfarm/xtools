package gocraft

import (
	"bytes"
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/gojek/work"
	"github.com/gojekfarm/xtools/xworker"
)

const payloadByteKey = "payload"
const payloadRawKey = "payload_raw"

var mapType = reflect.TypeOf(map[string]interface{}{})

// Enqueue enqueues an xworker.Job with xworker.Option(s).
func (f *Fulfiller) Enqueue(_ context.Context, j *xworker.Job, opts ...xworker.Option) (*xworker.EnqueueResult, error) {
	if f.disableCodec {
		return f.enqueueWithoutCodec(j, opts)
	}

	b, err := f.encodeJobPayload(j)
	if err != nil {
		return nil, err
	}

	return f.enqueueWithOptions(j.Name, f.constructArgs(b), mapOptions(opts))
}

// EnqueuePeriodically schedules an xworker.Job with given cronSchedule and xworker.Option(s).
func (f *Fulfiller) EnqueuePeriodically(cronSchedule string, j *xworker.Job, _ ...xworker.Option) error {
	f.pool.PeriodicallyEnqueue(cronSchedule, j.Name)

	return nil
}

func (f *Fulfiller) enqueueWithoutCodec(j *xworker.Job, opts []xworker.Option) (*xworker.EnqueueResult, error) {
	args, ok := j.Payload.(map[string]interface{})
	if !ok && j.Payload != nil {
		return f.enqueueAfterTypeConvert(j, opts)
	}

	return f.enqueueWithOptions(j.Name, args, mapOptions(opts))
}

func (f *Fulfiller) enqueueAfterTypeConvert(j *xworker.Job, opts []xworker.Option) (*xworker.EnqueueResult, error) {
	p := reflect.ValueOf(j.Payload)
	if p.Kind() == reflect.Ptr {
		p = p.Elem()
	}

	if p.Type().ConvertibleTo(mapType) {
		args, _ := p.Convert(mapType).Interface().(map[string]interface{})

		return f.enqueueWithOptions(j.Name, args, mapOptions(opts))
	}

	return nil, fmt.Errorf("payload %s must be a type alias to map[string]interface{}", reflect.TypeOf(j.Payload))
}

func (f *Fulfiller) enqueueWithOptions(
	jobName string,
	args map[string]interface{},
	eo *enqueueOptions,
) (*xworker.EnqueueResult, error) {
	if eo.in != 0 {
		secondsFromNow := int64(eo.in.Seconds())

		sw, err := f.enqueueIn(jobName, secondsFromNow, args, eo)
		if err != nil {
			return nil, err
		}

		return f.scheduledJobResult(sw), nil
	}

	w, err := f.enqueue(jobName, args, eo)
	if err != nil {
		return nil, err
	}

	return f.jobResult(w), nil
}

func (f *Fulfiller) enqueue(
	jobName string,
	args map[string]interface{},
	eo *enqueueOptions,
) (*work.Job, error) {
	if eo.unique {
		return f.enqueuer.EnqueueUnique(jobName, args)
	}

	return f.enqueuer.Enqueue(jobName, args)
}

func (f *Fulfiller) enqueueIn(
	jobName string,
	secondsFromNow int64,
	args map[string]interface{},
	eo *enqueueOptions,
) (*work.ScheduledJob, error) {
	if eo.unique {
		return f.enqueuer.EnqueueUniqueIn(jobName, secondsFromNow, args)
	}

	return f.enqueuer.EnqueueIn(jobName, secondsFromNow, args)
}

func (f *Fulfiller) scheduledJobResult(sw *work.ScheduledJob) *xworker.EnqueueResult {
	if sw == nil {
		return xworker.NewEnqueueResult("work.ScheduledJob(nil)", nil)
	}

	return xworker.NewEnqueueResult(fmt.Sprintf("work.ScheduledJob(ID: %s, Name: %s, EnqueuedAt: %s, RunAt: %s)",
		sw.ID, sw.Name, time.Unix(sw.EnqueuedAt, 0).UTC(), time.Unix(sw.RunAt, 0).UTC()), sw)
}

func (f *Fulfiller) jobResult(w *work.Job) *xworker.EnqueueResult {
	if w == nil {
		return xworker.NewEnqueueResult("work.Job(nil)", nil)
	}

	return xworker.NewEnqueueResult(fmt.Sprintf("work.Job(ID: %s, Name: %s, EnqueuedAt: %s)",
		w.ID, w.Name, time.Unix(w.EnqueuedAt, 0).UTC()), w)
}

func (f *Fulfiller) encodeJobPayload(j *xworker.Job) ([]byte, error) {
	var buf bytes.Buffer
	if _, err := buf.ReadFrom(j); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

func (f *Fulfiller) constructArgs(b []byte) map[string]interface{} {
	if f.useRawEncodedPayload {
		return map[string]interface{}{
			payloadRawKey: rawPayload(b),
		}
	}

	return map[string]interface{}{
		payloadByteKey: b,
	}
}

// enqueuer helps to mock work.Enqueuer.
type enqueuer interface {
	Enqueue(jobName string, args map[string]interface{}) (*work.Job, error)
	EnqueueUnique(jobName string, args map[string]interface{}) (*work.Job, error)
	EnqueueIn(jobName string, secondsFromNow int64, args map[string]interface{}) (*work.ScheduledJob, error)
	EnqueueUniqueIn(jobName string, secondsFromNow int64, args map[string]interface{}) (*work.ScheduledJob, error)
}

type enqueueOptions struct {
	unique bool
	in     time.Duration
}

func mapOptions(opts []xworker.Option) *enqueueOptions {
	eo := &enqueueOptions{}

	for _, opt := range opts {
		switch opt.Type() {
		case xworker.UniqueOpt:
			eo.unique = true
		case xworker.InOpt:
			eo.in, _ = opt.Value().(time.Duration)
		case xworker.AtOpt:
			eo.in = time.Until(opt.Value().(time.Time))
		}
	}

	return eo
}
