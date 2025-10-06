package xapi

import (
	"context"
	"encoding/json"
	"net/http"
)

// EndpointHandler defines the interface for handling endpoint requests.
type EndpointHandler[TReq, TRes any] interface {
	Handle(ctx context.Context, req *TReq) (*TRes, error)
}

// EndpointFunc is a function type that implements EndpointHandler.
type EndpointFunc[TReq, TRes any] func(ctx context.Context, req *TReq) (*TRes, error)

// Handle implements the EndpointHandler interface.
func (e EndpointFunc[TReq, TRes]) Handle(ctx context.Context, req *TReq) (*TRes, error) {
	return e(ctx, req)
}

// Endpoint represents a type-safe HTTP endpoint with middleware and error handling.
type Endpoint[TReq, TRes any] struct {
	handler EndpointHandler[TReq, TRes]
	opts    *options
}

// NewEndpoint creates a new Endpoint with the given handler and options.
func NewEndpoint[TReq, TRes any](handler EndpointHandler[TReq, TRes], opts ...EndpointOption) *Endpoint[TReq, TRes] {
	e := &Endpoint[TReq, TRes]{
		handler: handler,
		opts: &options{
			middleware:   MiddlewareStack{},
			errorHandler: ErrorFunc(DefaultErrorHandler),
		},
	}

	for _, option := range opts {
		option.apply(e.opts)
	}

	return e
}

// Handler returns an http.Handler that processes requests for this endpoint.
func (e *Endpoint[TReq, TRes]) Handler() http.Handler {
	h := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req TReq
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			e.opts.errorHandler.HandleError(w, err)
			return
		}

		res, err := e.handler.Handle(r.Context(), &req)
		if err != nil {
			e.opts.errorHandler.HandleError(w, err)
			return
		}

		if err := json.NewEncoder(w).Encode(res); err != nil {
			e.opts.errorHandler.HandleError(w, err)
			return
		}
	})

	if len(e.opts.middleware) > 0 {
		return e.opts.middleware.Middleware(h)
	}

	return h
}
