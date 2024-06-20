package xpod

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
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
	HealthCheckers []Checker
	ReadyCheckers  []Checker
	BuildInfo      *BuildInfo

	// Prefix is the base path for the health, ready, and version endpoints.
	// If not provided, the default value is "/".
	// If the prefix is "/probe", the health, ready, and version endpoints
	// will be available at:
	// - /probe/healthz
	// - /probe/readyz
	// - /probe/version
	//
	// HealthPath is the path for the health endpoint.
	// If not provided, the default value is "healthz".
	//
	// ReadyPath is the path for the readiness endpoint.
	// If not provided, the default value is "readyz".
	//
	// VersionPath is the path for the version endpoint.
	// If not provided, the default value is "version".
	//
	// Note: Prefix, HealthPath, ReadyPath, and VersionPath are only used
	// for internal http.ServeMux registration.
	Prefix, HealthPath, ReadyPath, VersionPath string

	ErrorLogDelegate func(string, map[string]any)

	// ShowErrReasons is used to show the error reasons in the HTTP response.
	ShowErrReasons bool
}

// NewProbeHandler returns a http.Handler which can be used to serve health check and build info endpoints.
func NewProbeHandler(opts Options) *ProbeHandler {
	ph := &ProbeHandler{
		sm:             http.NewServeMux(),
		logDelegate:    opts.ErrorLogDelegate,
		showErrReasons: opts.ShowErrReasons,
	}

	ph.makeHandlers(opts)
	ph.registerRoutes(opts)

	return ph
}

// ProbeHandler implements http.Handler interface to expose [/healthz /readyz /version] endpoints.
type ProbeHandler struct {
	sm             *http.ServeMux
	hh             http.Handler
	rh             http.Handler
	vh             http.Handler
	showErrReasons bool
	logDelegate    func(string, map[string]interface{})
}

func (h *ProbeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) { h.sm.ServeHTTP(w, r) }

// HealthHandler returns the handler for the health endpoint.
func (h *ProbeHandler) HealthHandler() http.Handler { return h.hh }

// ReadyHandler returns the handler for the readiness endpoint.
func (h *ProbeHandler) ReadyHandler() http.Handler { return h.rh }

// VersionHandler returns the handler for the version endpoint.
// Note: It will be nil if the Options.BuildInfo is not provided.
func (h *ProbeHandler) VersionHandler() http.Handler { return h.vh }

func (h *ProbeHandler) registerRoutes(opts Options) {
	prefix := strings.TrimSuffix(opts.Prefix, "/")

	h.sm.HandleFunc(
		fmt.Sprintf("%s/%s", prefix, pathOrDefault(opts.HealthPath, "healthz")),
		h.hh.ServeHTTP,
	)

	h.sm.HandleFunc(
		fmt.Sprintf("%s/%s", prefix, pathOrDefault(opts.ReadyPath, "readyz")),
		h.rh.ServeHTTP,
	)

	if h.vh != nil {
		h.sm.HandleFunc(
			fmt.Sprintf("%s/%s", prefix, pathOrDefault(opts.HealthPath, "version")),
			h.vh.ServeHTTP,
		)
	}
}

func pathOrDefault(path, def string) string {
	if strings.TrimSpace(path) != "" {
		return strings.TrimPrefix(path, "/")
	}

	return def
}

func (h *ProbeHandler) healthHandler(opts Options) http.Handler {
	hcs := opts.HealthCheckers
	if len(hcs) == 0 {
		hcs = append(hcs, PingHealthz)
	}

	return h.serveCheckers(hcs)
}

func (h *ProbeHandler) readyHandler(opts Options) http.Handler {
	rcs := opts.ReadyCheckers
	if len(rcs) == 0 {
		rcs = append(rcs, PingHealthz)
	}

	return h.serveCheckers(rcs)
}

func (h *ProbeHandler) serveCheckers(cs []Checker) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var excluded generic.Set[string]
		if reqExcludes, ok := r.URL.Query()[excludeQueryParam]; ok && len(reqExcludes) > 0 {
			excluded = generic.NewSet(flattenElems(slice.Map(r.URL.Query()[excludeQueryParam],
				func(s string) []string { return strings.Split(s, ",") },
			))...)
		}

		var output bytes.Buffer
		var failedVerboseLogOutput bytes.Buffer
		var failedChecks []*checkFailure

		for _, c := range cs {
			if excluded.Has(c.Name()) {
				excluded.Delete(c.Name())
				_, _ = fmt.Fprintf(&output, "[+]%s excluded: ok\n", c.Name())

				continue
			}

			if err := c.Check(r); err != nil {
				_, _ = fmt.Fprintf(&output, "[-]%s failed:", c.Name())

				if h.showErrReasons {
					_, _ = fmt.Fprintf(&output, "\n\treason: %v\n", err)
				} else {
					_, _ = fmt.Fprintf(&output, " reason hidden\n")
				}

				failedChecks = append(failedChecks, &checkFailure{name: c.Name(), err: err})
				_, _ = fmt.Fprintf(&failedVerboseLogOutput, "[-]%s failed: %v\n", c.Name(), err)

				continue
			}

			_, _ = fmt.Fprintf(&output, "[+]%s ok\n", c.Name())
		}

		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")

		if excluded.Len() > 0 {
			quotedChecks := strings.Join(
				slice.Map(excluded.Values(), func(in string) string { return fmt.Sprintf("%q", in) }), ", ")

			_, _ = fmt.Fprintf(&output, "warn: some checks cannot be excluded: no matches for %s\n", quotedChecks)
			if h.logDelegate != nil {
				h.logDelegate("cannot exclude some checks", map[string]interface{}{
					"checks": quotedChecks,
					"reason": "no matches",
				})
			}
		}

		checkPath := strings.TrimPrefix(r.URL.Path, "/")

		if len(failedChecks) > 0 {
			if h.logDelegate != nil {
				h.logDelegate(fmt.Sprintf("%s check failed", checkPath), map[string]interface{}{
					"failed_checks": strings.Join(
						slice.Map(failedChecks, func(cf *checkFailure) string { return cf.name }), ", "),
					"errs": errors.Join(slice.Map(failedChecks, func(cf *checkFailure) error { return cf.err })...),
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
		_, _ = fmt.Fprintf(w, "%s check passed\n", checkPath)
	})
}

func (h *ProbeHandler) makeHandlers(opts Options) {
	h.hh = h.healthHandler(opts)
	h.rh = h.readyHandler(opts)

	if opts.BuildInfo != nil {
		h.vh = h.versionHandler(opts)
	}
}

type checkFailure struct {
	name string
	err  error
}

func flattenElems(in [][]string) []string {
	var out []string

	for _, v := range in {
		out = append(out, v...)
	}

	return out
}
