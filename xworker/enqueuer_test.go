package xworker

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAdapter_EnqueuePeriodically(t *testing.T) {
	tests := []struct {
		name         string
		cronSchedule string
		job          *Job
		options      []Option
		wantErr      bool
	}{
		{
			name:         "Success",
			cronSchedule: "* * * * *",
			job:          &Job{Name: "test-job"},
		},
		{
			name:         "Error",
			cronSchedule: "* * * * *",
			job:          &Job{Name: "test-job"},
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mf := newMockFulfiller(t)
			if tt.wantErr {
				mf.On("EnqueuePeriodically", tt.cronSchedule, tt.job, tt.options).
					Return(errors.New("enqueue error"))
			} else {
				mf.On("EnqueuePeriodically", tt.cronSchedule, tt.job, tt.options).
					Return(nil)
			}

			a := &Adapter{fulfiller: mf}
			if err := a.EnqueuePeriodically(tt.cronSchedule, tt.job, tt.options...); tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestEnqueuerFunc_Enqueue(t *testing.T) {
	type args struct {
		ctx     context.Context
		j       *Job
		options []Option
	}
	tests := []struct {
		name    string
		f       EnqueuerFunc
		args    args
		want    *EnqueueResult
		wantErr bool
	}{
		{
			name: "Success",
			f: func(ctx context.Context, j *Job, option ...Option) (*EnqueueResult, error) {
				assert.Equal(t, "test-job", j.Name)
				return NewEnqueueResult("EnqueueResult", nil), nil
			},
			args: args{
				j: &Job{Name: "test-job"},
			},
			want: &EnqueueResult{str: "EnqueueResult"},
		},
		{
			name: "Error",
			f: func(ctx context.Context, j *Job, option ...Option) (*EnqueueResult, error) {
				assert.Equal(t, "test-job", j.Name)
				return nil, errors.New("enqueue error")
			},
			args: args{
				j: &Job{Name: "test-job"},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.f.Enqueue(tt.args.ctx, tt.args.j, tt.args.options...)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			assert.Equal(t, tt.want, got)
		})
	}
}

func TestPeriodicEnqueuerFunc_EnqueuePeriodically(t *testing.T) {
	tests := []struct {
		name         string
		f            PeriodicEnqueuerFunc
		cronSchedule string
		job          *Job
		options      []Option
		wantErr      bool
	}{
		{
			name: "Success",
			f: func(cronSchedule string, j *Job, option ...Option) error {
				assert.Equal(t, "* * * * *", cronSchedule)
				assert.Equal(t, "test-job", j.Name)
				return nil
			},
			cronSchedule: "* * * * *",
			job:          &Job{Name: "test-job"},
		},
		{
			name: "Error",
			f: func(cronSchedule string, j *Job, option ...Option) error {
				assert.Equal(t, "* b * * *", cronSchedule)
				assert.Equal(t, "test-job", j.Name)
				return errors.New("cron parse error")
			},
			cronSchedule: "* b * * *",
			job:          &Job{Name: "test-job"},
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.f.EnqueuePeriodically(tt.cronSchedule, tt.job, tt.options...); tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
