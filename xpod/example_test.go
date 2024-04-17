package xpod_test

import (
	"net/http"
	"time"

	"github.com/gojekfarm/xtools/xpod"
)

func ExampleNew() {
	h := http.NewServeMux()

	// This method automatically registers the health, ready, and version endpoints
	// with the provided prefix.
	// If you want to register the endpoints with a custom path, you can use the
	// `HealthHandler`, `ReadyHandler`, and `VersionHandler` methods. See the examples below.
	h.Handle("/probe", xpod.New(xpod.Options{
		Prefix: "/probe",
		BuildInfo: &xpod.BuildInfo{
			Version:   "0.1.0",
			Tag:       "v0.1.0",
			Commit:    "24b3f5d876ffa402287bfa5c26cf05626a2b3b01",
			BuildDate: xpod.BuildDate(time.Now()),
		},
	}))
}

func ExampleNew_withoutManagedServeMux() {
	h := http.NewServeMux()

	ph := xpod.New(xpod.Options{
		Prefix: "/probe",
		BuildInfo: &xpod.BuildInfo{
			Version:   "0.1.0",
			Tag:       "v0.1.0",
			Commit:    "24b3f5d876ffa402287bfa5c26cf05626a2b3b01",
			BuildDate: xpod.BuildDate(time.Now()),
		},
	})

	h.Handle("/health", ph.HealthHandler())
	h.Handle("/ready", ph.ReadyHandler())
	h.Handle("/version", ph.VersionHandler())
}
