package xpod

import (
	"bytes"
	"fmt"
	"net/http"
	"runtime"
	"strings"

	"github.com/gojekfarm/xtools/generic"
	"github.com/gojekfarm/xtools/generic/slice"
)

const (
	verboseQueryParam = "verbose"
	excludeQueryParam = "exclude"
)

// Options can be used to provide custom health/readiness checkers and the current BuildInfo.
type Options struct {
	Prefix           string
	HealthCheckers   []HealthChecker
	ReadyCheckers    []HealthChecker
	BuildInfo        *BuildInfo
	ErrorLogDelegate func(string, map[string]interface{})
	ShowErrReasons   bool
}

// NewProbeHandler returns a http.Handler which can be used to serve health check and build info endpoints.
func NewProbeHandler(opts Options) *ProbeHandler {
	ph := &ProbeHandler{sm: http.NewServeMux(), bi: &buildInfo{
		BuildInfo: opts.BuildInfo,
		GoVersion: runtime.Version(),
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	}, logDelegate: opts.ErrorLogDelegate, showErrReasons: opts.ShowErrReasons}

	ph.registerRoutes(strings.TrimSuffix(opts.Prefix, "/"), opts.HealthCheckers, opts.ReadyCheckers)

	return ph
}

// ProbeHandler implements http.Handler interface to expose [/healthz /readyz /version] endpoints.
type ProbeHandler struct {
	sm             *http.ServeMux
	bi             *buildInfo
	showErrReasons bool
	logDelegate    func(string, map[string]interface{})
}

func (h *ProbeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) { h.sm.ServeHTTP(w, r) }

func (h *ProbeHandler) registerRoutes(prefix string, hcs []HealthChecker, rcs []HealthChecker) {
	if len(hcs) == 0 {
		hcs = append(hcs, PingHealthz)
	}

	h.sm.HandleFunc(prefix+"/healthz", h.serveHealth(hcs).ServeHTTP)

	if len(rcs) == 0 {
		rcs = append(rcs, PingHealthz)
	}

	h.sm.HandleFunc(prefix+"/readyz", h.serveHealth(rcs).ServeHTTP)

	if h.bi.BuildInfo != nil {
		h.sm.HandleFunc(prefix+"/version", h.serveBuildInfo)
	}
}

func (h *ProbeHandler) serveHealth(hcs []HealthChecker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var excluded generic.Set[string]
		if reqExcludes, ok := r.URL.Query()[excludeQueryParam]; ok && len(reqExcludes) > 0 {
			excluded = generic.NewSet(flattenElems(slice.Map(
				r.URL.Query()[excludeQueryParam],
				func(s string) []string { return strings.Split(s, ",") },
			))...)
		}

		var output bytes.Buffer
		var failedVerboseLogOutput bytes.Buffer
		var failedChecks []string

		for _, hc := range hcs {
			if excluded.Has(hc.Name()) {
				excluded.Delete(hc.Name())
				_, _ = fmt.Fprintf(&output, "[+]%s excluded: ok\n", hc.Name())

				continue
			}

			if err := hc.Check(r); err != nil {
				_, _ = fmt.Fprintf(&output, "[-]%s failed:", hc.Name())

				if h.showErrReasons {
					_, _ = fmt.Fprintf(&output, "\n\treason: %v\n", err)
				} else {
					_, _ = fmt.Fprintf(&output, " reason hidden\n")
				}

				failedChecks = append(failedChecks, hc.Name())
				_, _ = fmt.Fprintf(&failedVerboseLogOutput, "[-]%s failed: %v\n", hc.Name(), err)

				continue
			}

			_, _ = fmt.Fprintf(&output, "[+]%s ok\n", hc.Name())
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")

		if excluded.Len() > 0 {
			quotedChecks := strings.Join(
				slice.Map(excluded.Values(),
					func(in string) string {
						return fmt.Sprintf("%q", in)
					}), ", ")

			_, _ = fmt.Fprintf(&output, "warn: some health checks cannot be excluded: no matches for %s\n", quotedChecks)
			if h.logDelegate != nil {
				h.logDelegate("cannot exclude some health checks", map[string]interface{}{
					"checks": quotedChecks,
					"reason": "no matches",
				})
			}
		}

		if len(failedChecks) > 0 {
			if h.logDelegate != nil {
				h.logDelegate("health check failed", map[string]interface{}{
					"failed_checks": strings.Join(failedChecks, ","),
				})
			}

			w.WriteHeader(http.StatusInternalServerError)
			_, _ = output.WriteTo(w)

			return
		}

		if _, found := r.URL.Query()[verboseQueryParam]; !found {
			_, _ = fmt.Fprint(w, "ok")

			return
		}

		_, _ = output.WriteTo(w)
		_, _ = fmt.Fprintf(w, "%s check passed\n", strings.TrimPrefix(r.URL.Path, "/"))
	})
}

func flattenElems(in [][]string) []string {
	var out []string

	for _, v := range in {
		out = append(out, v...)
	}

	return out
}
