package xpod_test

import (
	"net/http"
	"time"

	"github.com/gojekfarm/xtools/xpod"
)

func ExampleProbeHandler() {
	h := http.NewServeMux()
	h.Handle("/probe", xpod.NewProbeHandler(xpod.Options{
		BuildInfo: &xpod.BuildInfo{
			Version:   "0.1.0",
			Tag:       "v0.1.0",
			Commit:    "24b3f5d876ffa402287bfa5c26cf05626a2b3b01",
			BuildDate: xpod.BuildDate(time.Now()),
		},
	}))
}
