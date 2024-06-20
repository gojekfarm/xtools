package gocraft

import (
	"github.com/gojek/work"
	"github.com/gojekfarm/xtools/xworker"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestFulfiller_backoffCalculatorFunc(t *testing.T) {
	tests := []struct {
		name string
		opts xworker.RegisterOptions
		want []int64
	}{
		{
			name: "ConstRetryBackoffStrategy",
			opts: xworker.RegisterOptions{RetryBackoffStrategy: xworker.ConstantRetryBackoff(10 * time.Second)},
			want: []int64{10, 10, 10},
		},
		{
			name: "LinearRetryBackoffStrategy",
			opts: xworker.RegisterOptions{
				RetryBackoffStrategy: xworker.LinearRetryBackoff{
					InitialDelay: 10 * time.Second,
					Step:         10 * time.Second,
				},
			},
			want: []int64{10, 20, 30},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			f := &Fulfiller{disableCodec: true}

			bf := f.backoffCalculatorFunc(tt.opts)

			j := &work.Job{}
			for i, it := range tt.want {
				j.Fails = int64(i + 1)
				assert.Equal(t, it, bf(j))
			}
		})
	}
}
