package xpod

import (
	"net/http"
)

// HealthChecker is a named healthz checker.
type HealthChecker interface {
	Name() string
	Check(r *http.Request) error
}

// HealthCheckerFunc implements HealthChecker interface.
func HealthCheckerFunc(name string, check func(*http.Request) error) *HealthCheckerFun {
	return &HealthCheckerFun{name: name, checker: check}
}

// HealthCheckerFun implements HealthChecker interface.
type HealthCheckerFun struct {
	name    string
	checker func(r *http.Request) error
}

// Name returns the name of the health check.
func (f *HealthCheckerFun) Name() string { return f.name }

// Check is used to invoke health check when a request is received.
func (f *HealthCheckerFun) Check(req *http.Request) error { return f.checker(req) }

// PingHealthz returns true automatically when checked.
var PingHealthz HealthChecker = ping{}

// ping implements the simplest possible healthz checker.
type ping struct{}

func (ping) Name() string { return "ping" }

// Check is a health check that returns true.
func (ping) Check(_ *http.Request) error { return nil }
