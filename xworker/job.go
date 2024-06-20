package xworker

import (
	"bytes"
	"context"
	"io"
	"sync"
)

// Handler defines the handling of a Job.
type Handler interface {
	// Handle is the func signature for handling a Job.
	Handle(context.Context, *Job) error
}

// HandlerFunc is helper for creating a Handler.
type HandlerFunc func(context.Context, *Job) error

// Handle is the func signature for handling a Job.
func (f HandlerFunc) Handle(ctx context.Context, j *Job) error {
	return f(ctx, j)
}

// Job denotes a task with a Name and Payload to pass to workers.
type Job struct {
	// Name represents the name of the job/task.
	Name string `json:"name,omitempty"`
	// Payload holds the object that is to be passed to worker.
	//
	// When enqueueing, it can contain any value, it will be converted to bytes by using the configured PayloadEncoder.
	//
	// When reading in Handler, its value is nil, use DecodePayload method to convert back to runtime object.
	Payload interface{} `json:"payload,omitempty"`
	// Raw holds the value for actual type that implemented the job/task.
	Raw interface{} `json:"-"`

	once        sync.Once
	buf         bytes.Buffer
	encoderFunc PayloadEncoderFunc
	decoderFunc PayloadDecoderFunc
}

// NewJobWithDecoder returns a Job with a PayloadDecoderFunc set.
func NewJobWithDecoder(decoder PayloadDecoderFunc) *Job {
	return &Job{decoderFunc: decoder}
}

func (j *Job) Write(p []byte) (n int, err error) {
	j.buf.Reset()

	return j.buf.Write(p)
}

func (j *Job) Read(p []byte) (int, error) {
	var e error

	j.once.Do(func() {
		if err := j.encoderFunc(&j.buf).Encode(j.Payload); err != nil {
			e = err
		}
	})

	if e != nil {
		return 0, e
	}

	n, e := j.buf.Read(p)

	if j.buf.Len() == 0 {
		return n, io.EOF
	}

	return n, e
}

// DecodePayload uses AdapterOptions.PayloadDecoder to decode the job payload bytes into v.
func (j *Job) DecodePayload(v interface{}) error {
	return j.decoderFunc(bytes.NewBuffer(j.buf.Bytes())).Decode(v)
}
