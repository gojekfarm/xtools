package xpod

import (
	"net/http"
)

// Checker is a named resource/component checker.
type Checker interface {
	Name() string
	Check(r *http.Request) error
}

// CheckerFunc implements Checker interface.
func CheckerFunc(name string, check func(*http.Request) error) *CheckerFun {
	return &CheckerFun{name: name, checker: check}
}

// CheckerFun implements Checker interface.
type CheckerFun struct {
	name    string
	checker func(r *http.Request) error
}

// Name returns the name of the health check.
func (f *CheckerFun) Name() string { return f.name }

// Check is used to invoke health check when a request is received.
func (f *CheckerFun) Check(req *http.Request) error { return f.checker(req) }

// PingHealthz returns true automatically when checked.
var PingHealthz Checker = ping{}

// ping implements the simplest possible healthz checker.
type ping struct{}

func (ping) Name() string { return "ping" }

// Check is a health check that returns true.
func (ping) Check(_ *http.Request) error { return nil }
