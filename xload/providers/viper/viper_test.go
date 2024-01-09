package viper

import (
	"context"
	"os"
	"testing"

	vpr "github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	testcases := []struct {
		name            string
		tempFileContent string
		opts            func(tempFile string) []Option
		wantErr         assert.ErrorAssertionFunc
		assertion       func(t *testing.T, l *Loader)
	}{
		{
			name:    "default",
			wantErr: assert.NoError,
			assertion: func(t *testing.T, l *Loader) {
				assert.NotNil(t, l)
			},
		},
		{
			name: "with config file",
			tempFileContent: `
foo:
  bar: baz
`,
			opts: func(tempFile string) []Option {
				return []Option{ConfigFile(tempFile)}
			},
			wantErr: assert.NoError,
			assertion: func(t *testing.T, l *Loader) {
				assert.NotNil(t, l)
				t.Logf("config file used: %s", l.ConfigFileUsed())
				v, err := l.Load(context.TODO(), "foo_bar")
				assert.NoError(t, err)
				assert.Equal(t, "baz", v)
			},
		},
	}

	for _, tc := range testcases {
		t.Run(tc.name, func(t *testing.T) {
			var opts []Option

			if tc.opts != nil {
				opts = tc.opts("")
			}

			if tc.tempFileContent != "" {
				f, err := os.CreateTemp("", "config_*.yaml")
				assert.NoError(t, err)

				_, err = f.WriteString(tc.tempFileContent)
				assert.NoError(t, err)

				opts = tc.opts(f.Name())

				defer func() {
					assert.NoError(t, f.Close())
					assert.NoError(t, os.Remove(f.Name()))
				}()
			}

			l, err := New(opts...)
			tc.wantErr(t, err)
			tc.assertion(t, l)
		})
	}

	t.Run("unsupported config type", func(t *testing.T) {
		_, err := New(ConfigFile("foo.bar"))
		var vErr vpr.UnsupportedConfigError
		assert.ErrorAs(t, err, &vErr)
	})
}
