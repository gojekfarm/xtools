package xpod

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/gojekfarm/xtools/generic/slice"
)

func TestProbeHandler_serveHealth(t *testing.T) {
	tests := []struct {
		name        string
		opts        Options
		verbose     bool
		excluded    []string
		logDelegate func(*testing.T, *mock.Mock)
		want        string
	}{
		{
			name: "NoHealthCheckNonVerbose",
			want: "ok",
		},
		{
			name:    "NoHealthCheckVerbose",
			verbose: true,
			want: `[+]ping ok
healthz check passed
`,
		},
		{
			name: "FailingHealthCheckWithHiddenReason",
			opts: Options{
				HealthCheckers: []HealthChecker{
					HealthCheckerFunc("redis", func(_ *http.Request) error {
						return errors.New("redis-connect-error")
					}),
				},
			},
			logDelegate: func(t *testing.T, m *mock.Mock) {
				m.On("Logf", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					argsMap := args.Get(1).(map[string]interface{})

					assert.Equal(t, "health check failed", args.String(0))
					assert.Equal(t, "redis", argsMap["failed_checks"])
				})
			},
			want: `[-]redis failed: reason hidden
`,
		},
		{
			name: "FailingHealthCheckWithReason",
			opts: Options{
				HealthCheckers: []HealthChecker{
					HealthCheckerFunc("redis", func(_ *http.Request) error {
						return errors.New("redis-connect-error")
					}),
				},
				ShowErrReasons: true,
			},
			want: `[-]redis failed:
	reason: redis-connect-error
`,
		},
		{
			name: "FailingHealthCheckExcluded",
			opts: Options{
				HealthCheckers: []HealthChecker{
					HealthCheckerFunc("redis", func(_ *http.Request) error {
						return errors.New("redis-connect-error")
					}),
					PingHealthz,
				},
			},
			excluded: []string{"redis"},
			want:     "ok",
		},
		{
			name:     "FailingHealthCheckWithExtraExcludes",
			verbose:  true,
			excluded: []string{"redis", "foo,bar", "baz"},
			opts: Options{
				HealthCheckers: []HealthChecker{
					PingHealthz,
					HealthCheckerFunc("redis", func(_ *http.Request) error {
						return errors.New("redis-connect-error")
					}),
				},
			},
			logDelegate: func(t *testing.T, m *mock.Mock) {
				m.On("Logf", mock.Anything, mock.Anything).Run(func(args mock.Arguments) {
					argsMap := args.Get(1).(map[string]interface{})

					assert.Equal(t, "cannot exclude some health checks", args.String(0))
					assert.ElementsMatch(t, []string{"foo", "bar", "baz"}, slice.Map(
						strings.Split(argsMap["checks"].(string), ", "), func(s string) string {
							return strings.Trim(s, `"`)
						},
					))
					assert.Equal(t, "no matches", argsMap["reason"])
				})
			},
			want: `[+]ping ok
[+]redis excluded: ok
warn: some health checks cannot be excluded: no matches for "foo", "bar", "baz"
healthz check passed
`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ld := newMockLogDelegate(t)

			path := tt.opts.Prefix + "/healthz?"

			if len(tt.excluded) > 0 {
				for _, e := range tt.excluded {
					path += "exclude=" + e + "&"
				}
			}

			if tt.verbose {
				path += "verbose"
			}

			req := httptest.NewRequest(http.MethodGet, path, nil)
			rw := httptest.NewRecorder()

			if tt.logDelegate != nil {
				tt.logDelegate(t, &ld.Mock)
				tt.opts.ErrorLogDelegate = ld.Logf
			}

			NewProbeHandler(tt.opts).ServeHTTP(rw, req)

			rc := rw.Result().Body
			b, err := io.ReadAll(rc)

			assert.NoError(t, err)
			assert.NoError(t, rc.Close())

			assert.Equal(t, tt.want, string(b))

			ld.AssertExpectations(t)
		})
	}
}

func newMockLogDelegate(t *testing.T) *mockLogDelegate {
	m := &mockLogDelegate{}
	m.Test(t)
	return m
}

type mockLogDelegate struct{ mock.Mock }

func (m *mockLogDelegate) Logf(format string, args map[string]interface{}) { m.Called(format, args) }
