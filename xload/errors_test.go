package xload

import (
	"errors"
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

func TestErrCollision(t *testing.T) {
	ks := []string{
		"KEY_A",
		"KEY_B",
	}
	err := &ErrCollision{keys: ks}

	assert.ElementsMatch(t, ks, err.Keys())
	assert.Equal(t, "xload: key collisions detected for keys: [KEY_A KEY_B]", err.Error())
}

func TestErrDecodeValue(t *testing.T) {
	errDecode := errors.New("decode error")
	err := &ErrDecode{
		key: "KEY_A",
		err: errDecode,
	}

	assert.EqualError(t, err, "xload: unable to decode value for key KEY_A: decode error")
	assert.ErrorIs(t, err, errDecode)
}
