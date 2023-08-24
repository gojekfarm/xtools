package xload

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_defaultOptions(t *testing.T) {
	want := &options{
		tagName:     defaultKey,
		loader:      OSLoader(),
		concurrency: 1,
	}
	opts := newOptions()

	assert.Equal(t, want.tagName, opts.tagName)
	assert.Equal(t, want.concurrency, opts.concurrency)
	assert.IsType(t, want.loader, opts.loader)
}

func TestOptions(t *testing.T) {
	testcases := []struct {
		name string
		opts []Option
		want *options
	}{
		{
			name: "field tag name",
			opts: []Option{FieldTagName("custom")},
			want: &options{
				tagName:     "custom",
				loader:      OSLoader(),
				concurrency: 1,
			},
		},
		{
			name: "loader",
			opts: []Option{MapLoader{"A": "1"}},
			want: &options{
				tagName:     defaultKey,
				loader:      MapLoader{"A": "1"},
				concurrency: 1,
			},
		},
		{
			name: "concurrency",
			opts: []Option{Concurrency(2)},
			want: &options{
				tagName:     defaultKey,
				loader:      OSLoader(),
				concurrency: 2,
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			opts := newOptions(tc.opts...)

			assert.Equal(t, tc.want.tagName, opts.tagName)
			assert.Equal(t, tc.want.concurrency, opts.concurrency)
			assert.IsType(t, tc.want.loader, opts.loader)
		})
	}
}
