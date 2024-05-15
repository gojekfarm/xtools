package xload

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_defaultOptions(t *testing.T) {
	want := &options{
		tagName:          defaultKey,
		loader:           OSLoader(),
		concurrency:      1,
		detectCollisions: true,
	}
	opts := newOptions()

	t.Run("Loader", func(t *testing.T) {
		assert.IsType(t, want.loader, opts.loader)
		want.loader = nil
		opts.loader = nil
	})

	assert.Equal(t, want, opts)
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
				tagName:          "custom",
				loader:           OSLoader(),
				concurrency:      1,
				detectCollisions: true,
			},
		},
		{
			name: "loader",
			opts: []Option{MapLoader{"A": "1"}},
			want: &options{
				tagName:          defaultKey,
				loader:           MapLoader{"A": "1"},
				concurrency:      1,
				detectCollisions: true,
			},
		},
		{
			name: "concurrency",
			opts: []Option{Concurrency(2)},
			want: &options{
				tagName:          defaultKey,
				loader:           OSLoader(),
				concurrency:      2,
				detectCollisions: true,
			},
		},
		{
			name: "detectCollisions",
			opts: []Option{SkipCollisionDetection},
			want: &options{
				tagName:     defaultKey,
				loader:      OSLoader(),
				concurrency: 1,
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			opts := newOptions(tc.opts...)

			t.Run("Loader", func(t *testing.T) {
				assert.IsType(t, tc.want.loader, opts.loader)
				tc.want.loader = nil
				opts.loader = nil
			})

			assert.Equal(t, tc.want, opts)
		})
	}
}
