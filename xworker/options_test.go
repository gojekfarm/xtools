package xworker

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestOptions(t *testing.T) {
	atTime := time.Date(2021, 10, 23, 10, 30, 00, 00, time.UTC)

	testcases := []struct {
		name     string
		option   Option
		want     interface{}
		str      string
		wantType OptionType
	}{
		{
			name:     "Unique",
			option:   Unique,
			want:     true,
			wantType: UniqueOpt,
			str:      "Unique",
		},
		{
			name:     "In",
			option:   In(time.Second),
			want:     time.Second,
			wantType: InOpt,
			str:      "In(1s)",
		},
		{
			name:     "At",
			option:   At(atTime),
			want:     atTime,
			wantType: AtOpt,
			str:      "At(2021-10-23T10:30:00Z)",
		},
	}

	for _, tt := range testcases {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.option.Value())
			assert.Equal(t, tt.wantType, tt.option.Type())
			assert.Equal(t, tt.str, tt.option.String())
		})
	}
}
