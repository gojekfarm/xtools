package xconfig

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParse(t *testing.T) {
	testcases := []struct {
		name  string
		input any
		want  []*Node
		err   error
	}{
		{
			name: "simple struct",
			input: &struct {
				Name string `env:"NAME"`
			}{},
			want: []*Node{
				{
					name: "NAME",
				},
			},
		},
		{
			name: "required field",
			input: &struct {
				Name string `env:"NAME,required"`
			}{},
			want: []*Node{
				{
					name:     "NAME",
					required: true,
				},
			},
		},
		{
			name: "nested struct with prefix",
			input: &struct {
				Database struct {
					Host string `env:"HOST"`
				} `env:",prefix=DB_"`
				Name string `env:"NAME"`
			}{},
			want: []*Node{
				{
					name: "DB_HOST",
				},
				{
					name: "NAME",
				},
			},
		},
		{
			name: "nested struct without prefix",
			input: &struct {
				Database struct {
					Host string `env:"HOST"`
				}
				Name string `env:"NAME"`
			}{},
			want: []*Node{
				{
					name: "HOST",
				},
				{
					name: "NAME",
				},
			},
		},
	}

	for _, tc := range testcases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			got, err := parse(tc.input, &options{key: "env"})

			if tc.err != nil {
				assert.Error(t, err)
				assert.ErrorIs(t, err, tc.err)

				return
			}

			assert.EqualValues(t, tc.want, got)
		})
	}
}