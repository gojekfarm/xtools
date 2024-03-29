package xload

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestErrUnknownTagOption_Error(t *testing.T) {
	tests := []struct {
		name string
		key  string
		opt  string
		want string
	}{
		{
			name: "key and opt",
			key:  "key",
			opt:  "opt",
			want: "`key` key has unknown tag option: opt",
		},
		{
			name: "opt only",
			opt:  "opt",
			want: "unknown tag option: opt",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := &ErrUnknownTagOption{
				key: tt.key,
				opt: tt.opt,
			}
			assert.Equalf(t, tt.want, e.Error(), "Error()")
		})
	}
}
