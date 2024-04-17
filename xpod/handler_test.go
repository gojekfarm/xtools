package xpod

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/gojekfarm/xtools/generic/slice"
)

func TestProbeHandler_serveCheckers(t *testing.T) {
	type args struct {
		name        string
		prefix      string
		excluded    []string
		verbose     bool
		want        func(*testing.T, string)
		logDelegate *mockLogDelegate
	}

	runSpecs := func(t *testing.T, handler http.Handler, a args) {
		t.Run(a.name, func(t *testing.T) {
			path := a.prefix + "/healthz?"

			if len(a.excluded) > 0 {
				for _, e := range a.excluded {
					path += "exclude=" + e + "&"
				}
			}

			if a.verbose {
				path += "verbose"
			}

			req := httptest.NewRequest(http.MethodGet, path, nil)
			rw := httptest.NewRecorder()

			handler.ServeHTTP(rw, req)

			rc := rw.Result().Body
			b, err := io.ReadAll(rc)

			assert.NoError(t, err)
			assert.NoError(t, rc.Close())

			a.want(t, string(b))
			a.logDelegate.AssertExpectations(t)
		})
	}

	tests := []struct {
		name        string
		opts        Options
		verbose     bool
		excluded    []string
		logDelegate func(*testing.T, *mock.Mock)
		want        func(*testing.T, string)
	}{
		{
			name: "NoHealthCheckNonVerbose",
			want: func(t *testing.T, got string) {
				assert.Equal(t, "ok", got)
			},
		},
		{
			name:    "NoHealthCheckVerbose",
			verbose: true,
			want: func(t *testing.T, got string) {
				assert.Equal(t, `[+]ping ok
healthz check passed
`, got)
			},
		},
		{
			name: "FailingHealthCheckWithHiddenReason",
			opts: Options{
				HealthCheckers: []HealthChecker{
					HealthCheckerFunc("redis", func(_ *http.Request) error {
						return errors.New("redis-connect-error")
					}),
				},
				ReadyCheckers: []HealthChecker{
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
			want: func(t *testing.T, got string) {
				assert.Equal(t, `[-]redis failed: reason hidden
`, got)
			},
		},
		{
			name: "FailingHealthCheckWithReason",
			opts: Options{
				HealthCheckers: []HealthChecker{
					HealthCheckerFunc("redis", func(_ *http.Request) error {
						return errors.New("redis-connect-error")
					}),
				},
				ReadyCheckers: []HealthChecker{
					HealthCheckerFunc("redis", func(_ *http.Request) error {
						return errors.New("redis-connect-error")
					}),
				},
				ShowErrReasons: true,
			},
			want: func(t *testing.T, got string) {
				assert.Equal(t, `[-]redis failed:
	reason: redis-connect-error
`, got)
			},
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
				ReadyCheckers: []HealthChecker{
					HealthCheckerFunc("redis", func(_ *http.Request) error {
						return errors.New("redis-connect-error")
					}),
					PingHealthz,
				},
			},
			excluded: []string{"redis"},
			want:     func(t *testing.T, got string) { assert.Equal(t, "ok", got) },
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
				ReadyCheckers: []HealthChecker{
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
			want: func(t *testing.T, got string) {
				assert.True(t, strings.Contains(got, `[+]ping ok`))
				assert.True(t, strings.Contains(got, `[+]redis excluded: ok`))

				re := regexp.MustCompile(`warn: some health checks cannot be excluded: no matches for (.*)`)
				assert.True(t, re.MatchString(got))

				excludedLine := re.FindStringSubmatch(got)[1]
				assert.ElementsMatch(t, []string{"foo", "bar", "baz"}, slice.Map(
					strings.Split(excludedLine, ", "), func(s string) string {
						return strings.Trim(s, `"`)
					},
				))
			},
		},
	}

	t.Run("New", func(t *testing.T) {
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				ld := newMockLogDelegate(t)
				if tt.logDelegate != nil {
					tt.logDelegate(t, &ld.Mock)
					tt.opts.ErrorLogDelegate = ld.Logf
				}

				handler := New(tt.opts)

				runSpecs(t, handler, args{
					name:        tt.name,
					prefix:      tt.opts.Prefix,
					excluded:    tt.excluded,
					verbose:     tt.verbose,
					want:        tt.want,
					logDelegate: ld,
				})
			})
		}
	})

	t.Run("HealthHandler", func(t *testing.T) {
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				ld := newMockLogDelegate(t)
				if tt.logDelegate != nil {
					tt.logDelegate(t, &ld.Mock)
					tt.opts.ErrorLogDelegate = ld.Logf
				}

				handler := New(tt.opts).HealthHandler()
				runSpecs(t, handler, args{
					name:        tt.name,
					prefix:      tt.opts.Prefix,
					excluded:    tt.excluded,
					verbose:     tt.verbose,
					want:        tt.want,
					logDelegate: ld,
				})
			})
		}
	})

	t.Run("ReadyHandler", func(t *testing.T) {
		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				ld := newMockLogDelegate(t)
				if tt.logDelegate != nil {
					tt.logDelegate(t, &ld.Mock)
					tt.opts.ErrorLogDelegate = ld.Logf
				}

				handler := New(tt.opts).ReadyHandler()
				runSpecs(t, handler, args{
					name:        tt.name,
					prefix:      tt.opts.Prefix,
					excluded:    tt.excluded,
					verbose:     tt.verbose,
					want:        tt.want,
					logDelegate: ld,
				})
			})
		}
	})
}

func newMockLogDelegate(t *testing.T) *mockLogDelegate {
	m := &mockLogDelegate{}
	m.Test(t)
	return m
}

type mockLogDelegate struct{ mock.Mock }

func (m *mockLogDelegate) Logf(format string, args map[string]interface{}) { m.Called(format, args) }

func Test_pathOrDefault(t *testing.T) {
	tests := []struct {
		name string
		path string
		def  string
		want string
	}{
		{
			name: "EmptyPath",
			def:  "healthz",
			want: "healthz",
		},
		{
			name: "EmptyDefault",
			path: "healthz",
			want: "healthz",
		},
		{
			name: "NonEmptyPath",
			path: "readyz",
			def:  "healthz",
			want: "readyz",
		},
		{
			name: "BothEmpty",
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, pathOrDefault(tt.path, tt.def), "pathOrDefault(%v, %v)", tt.path, tt.def)
		})
	}
}
