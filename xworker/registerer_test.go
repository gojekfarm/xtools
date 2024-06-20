package xworker

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestRegistererFunc_RegisterHandlerWithOptions(t *testing.T) {
	tests := []struct {
		name    string
		f       RegistererFunc
		wantErr bool
	}{
		{
			name: "Success",
			f: func(jobName string, jobHandler Handler, options RegisterOptions) error {
				return nil
			},
		},
		{
			name: "Error",
			f: func(jobName string, jobHandler Handler, options RegisterOptions) error {
				return errors.New("reg_err")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.f.RegisterHandlerWithOptions(
				"test-job",
				HandlerFunc(func(ctx context.Context, j *Job) error {
					return nil
				}),
				RegisterOptions{},
			)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestAdapter_RegisterHandlerWithOptions(t *testing.T) {
	mf := newMockFulfiller(t)
	mf.On(
		"RegisterHandlerWithOptions",
		"test-job",
		mock.AnythingOfType("xworker.HandlerFunc"),
		RegisterOptions{},
	).Return(nil).Once()

	a := Adapter{fulfiller: mf}

	assert.NoError(t, a.RegisterHandlerWithOptions("test-job",
		HandlerFunc(func(ctx context.Context, job *Job) error {
			return nil
		}),
		RegisterOptions{}),
	)

	mf.On(
		"RegisterHandlerWithOptions",
		"test-job",
		mock.AnythingOfType("xworker.HandlerFunc"),
		RegisterOptions{},
	).Return(errors.New("some_err")).Once()

	assert.Error(t, a.RegisterHandlerWithOptions("test-job",
		HandlerFunc(func(ctx context.Context, job *Job) error {
			return nil
		}),
		RegisterOptions{}),
	)
}

func TestAdapter_injectRetryBackoffWithDecoder(t *testing.T) {
	rbf := RetryBackoffFunc(func(_ int, _ error, _ *Job) time.Duration { return 0 })
	ro := &RegisterOptions{RetryBackoffStrategy: rbf}

	a := &Adapter{decoderFunc: DefaultPayloadDecoderFunc}
	a.injectRetryBackoffWithDecoder(ro)

	assert.NotEqual(t, fmt.Sprintf("%p", ro.RetryBackoffStrategy), fmt.Sprintf("%p", rbf))

	j := &Job{}
	ro.RetryBackoffStrategy.RetryBackoff(0, nil, j)

	assert.Equal(t, fmt.Sprintf("%p", j.decoderFunc), fmt.Sprintf("%p", a.decoderFunc))
}
