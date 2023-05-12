package xpod

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestProbeHandler_serveVersion(t *testing.T) {
	tests := []struct {
		name    string
		opts    Options
		verbose bool
		want    string
	}{
		{
			name: "NoVersionInfo",
			want: "404 page not found\n",
		},
		{
			name: "NonVerboseVersion",
			opts: Options{BuildInfo: &BuildInfo{Version: "0.1.0"}},
			want: "0.1.0",
		},
		{
			name:    "VerboseVersion",
			verbose: true,
			opts: Options{BuildInfo: &BuildInfo{
				Version: "0.1.0",
				Tag:     "v0.1.0",
				Commit:  "24b3f5d876ffa402287bfa5c26cf05626a2b3b01",
				BuildDate: BuildDate(
					time.Date(2022, 04, 20, 4, 20, 4, 20, time.UTC),
				),
			}},
			want: fmt.Sprintf(`{
  "version": "0.1.0",
  "tag": "v0.1.0",
  "commit": "24b3f5d876ffa402287bfa5c26cf05626a2b3b01",
  "build_date": "Wed, 20 Apr 2022 04:20:04 UTC",
  "go_version": "%s",
  "os": "%s",
  "arch": "%s"
}
`, runtime.Version(), runtime.GOOS, runtime.GOARCH),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			path := tt.opts.Prefix + "/version"
			if tt.verbose {
				path += "?verbose"
			}

			req := httptest.NewRequest(http.MethodGet, path, nil)
			rw := httptest.NewRecorder()

			NewProbeHandler(tt.opts).ServeHTTP(rw, req)

			b, err := io.ReadAll(rw.Result().Body)
			assert.NoError(t, err)

			assert.Equal(t, tt.want, string(b))
		})
	}
}
