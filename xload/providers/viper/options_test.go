package viper

import (
	"fmt"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"github.com/gojekfarm/xtools/xload"
)

func TestOptions(t *testing.T) {
	testcases := []struct {
		name string
		opts []Option
		want *options
	}{
		{
			name: "PrefixSeparator",
			opts: []Option{PrefixSeparator("__")},
			want: &options{separator: "__"},
		},
		{
			name: "ConfigFile",
			opts: []Option{ConfigFile("/tmp/config.yaml")},
			want: &options{file: fileOpts{absPath: "/tmp/config.yaml"}},
		},
		{
			name: "ConfigName",
			opts: []Option{ConfigName("config")},
			want: &options{file: fileOpts{name: "config"}},
		},
		{
			name: "ConfigType",
			opts: []Option{ConfigType("json")},
			want: &options{file: fileOpts{ext: "json"}},
		},
		{
			name: "ConfigPaths",
			opts: []Option{ConfigPaths{"/tmp", "/tmp/config"}},
			want: &options{file: fileOpts{paths: []string{"/tmp", "/tmp/config"}}},
		},
		{
			name: "AutoEnv",
			opts: []Option{AutoEnv(true)},
			want: &options{autoEnv: true},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			opts := &options{}
			for _, o := range tc.opts {
				o.apply(opts)
			}

			assert.Equal(t, tc.want, opts)
		})
	}

	t.Run("Transformer", func(t *testing.T) {
		opts := &options{}
		f := Transformer(func(_ *viper.Viper, _ xload.Loader) xload.Loader { return nil })
		f.apply(opts)
		assert.NotNil(t, opts.transform)
		assert.True(t, fmt.Sprintf("%p", f) == fmt.Sprintf("%p", opts.transform))
	})
}

func Test_def(t *testing.T) {
	opts := def()

	assert.Equal(t, fileOpts{name: "application", ext: "yaml", paths: []string{"./", "../"}}, opts.file)
	assert.Equal(t, "_", opts.separator)
	assert.NotNil(t, opts.viper)
	assert.NotNil(t, opts.transform)
}
