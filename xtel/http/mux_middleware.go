package http

import (
	"net/http"

	"github.com/gorilla/mux"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gorilla/mux/otelmux"
)

var defaultPathWhitelistFunc = PathWhitelistFunc(func(r *http.Request) bool { return true })

// MuxMiddleware sets up a handler to start tracing the incoming
// requests. The service parameter should describe the name of the
// (virtual) server handling the request.
func MuxMiddleware(service string, opts ...Option) mux.MiddlewareFunc {
	opt := newOptions(opts...)

	return func(next http.Handler) http.Handler {
		return muxWrapper{
			handler: next,
			wrappedHandler: otelmux.Middleware(service,
				otelmux.WithTracerProvider(opt.traceProvider),
				otelmux.WithPropagators(opt.propagator),
			).Middleware(next),
			wf: opt.pathWhitelistFunc,
		}
	}
}

type muxWrapper struct {
	handler        http.Handler
	wrappedHandler http.Handler
	wf             func(*http.Request) bool
}

// ServeHTTP is used to wrap the http.Handler with ServeHTTP.
func (mw muxWrapper) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if !mw.wf(r) {
		mw.handler.ServeHTTP(w, r)

		return
	}

	mw.wrappedHandler.ServeHTTP(w, r)
}
