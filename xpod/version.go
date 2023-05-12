package xpod

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// BuildDate is an alias to time.Time.
type BuildDate time.Time

func (d BuildDate) String() string {
	return time.Time(d).Format(time.RFC1123)
}

// MarshalJSON implements the json.Marshaler interface.
func (d BuildDate) MarshalJSON() ([]byte, error) {
	return []byte(fmt.Sprintf(`"%s"`, d.String())), nil
}

// BuildInfo holds the information about current build.
type BuildInfo struct {
	Version   string    `json:"version"`
	Tag       string    `json:"tag"`
	Commit    string    `json:"commit"`
	BuildDate BuildDate `json:"build_date"`
}

type buildInfo struct {
	*BuildInfo
	GoVersion string `json:"go_version"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
}

func (h *ProbeHandler) serveBuildInfo(w http.ResponseWriter, r *http.Request) {
	if _, found := r.URL.Query()[verboseQueryParam]; !found {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("X-Content-Type-Options", "nosniff")

		_, _ = fmt.Fprint(w, h.bi.Version)

		return
	}

	w.Header().Set("Content-Type", "application/json")

	e := json.NewEncoder(w)
	e.SetIndent("", "  ")
	_ = e.Encode(h.bi)
}
