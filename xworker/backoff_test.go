package xworker

import (
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func TestConstantRetryBackoff_RetryBackoff(t *testing.T) {
	tests := []struct {
		name string
		c    ConstantRetryBackoff
		want []time.Duration
	}{
		{
			name: "ZeroValue",
			want: []time.Duration{0, 0, 0},
		},
		{
			name: "ConstantValue",
			c:    ConstantRetryBackoff(10 * time.Second),
			want: []time.Duration{10 * time.Second},
		},
		{
			name: "MultipleConstantValues",
			c:    ConstantRetryBackoff(10 * time.Second),
			want: []time.Duration{10 * time.Second, 10 * time.Second, 10 * time.Second},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i, d := range tt.want {
				assert.Equal(t, d, tt.c.RetryBackoff(i, nil, nil))
			}
		})
	}
}

func TestLinearRetryBackoff_RetryBackoff(t *testing.T) {
	tests := []struct {
		name string
		l    LinearRetryBackoff
		want []time.Duration
	}{
		{
			name: "ZeroValue",
			want: []time.Duration{0, 0, 0},
		},
		{
			name: "ConstantValue",
			l:    LinearRetryBackoff{InitialDelay: 10 * time.Second},
			want: []time.Duration{10 * time.Second},
		},
		{
			name: "MultipleConstantValues",
			l:    LinearRetryBackoff{InitialDelay: 10 * time.Second, Step: 10 * time.Second},
			want: []time.Duration{10 * time.Second, 20 * time.Second, 30 * time.Second},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for i, d := range tt.want {
				assert.Equal(t, d, tt.l.RetryBackoff(i, nil, nil))
			}
		})
	}
}
