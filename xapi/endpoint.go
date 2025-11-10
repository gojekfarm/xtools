package xapi

import (
	"context"
	"encoding/json"
	"io"
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

// Extracter allows extracting additional data from the HTTP request,
// such as headers, query params, etc.
type Extracter interface {
	Extract(r *http.Request) error
}

// Validator allows validating endpoint requests.
type Validator interface {
	Validate() error
}

// StatusSetter allows setting a custom HTTP status code for the response.
type StatusSetter interface {
	StatusCode() int
}

// RawWriter allows writing raw data to the HTTP response instead of
// the default JSON encoder.
type RawWriter interface {
	Write(w http.ResponseWriter) error
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

		data, err := io.ReadAll(r.Body)
		if err != nil {
			e.opts.errorHandler.HandleError(w, err)
			return
		}
		defer r.Body.Close()

		if len(data) > 0 {
			if err := json.Unmarshal(data, &req); err != nil {
				e.opts.errorHandler.HandleError(w, err)
				return
			}
		}

		if extracter, ok := any(&req).(Extracter); ok {
			if err := extracter.Extract(r); err != nil {
				e.opts.errorHandler.HandleError(w, err)
				return
			}
		}

		if validator, ok := any(&req).(Validator); ok {
			if err := validator.Validate(); err != nil {
				e.opts.errorHandler.HandleError(w, err)
				return
			}
		}

		res, err := e.handler.Handle(r.Context(), &req)
		if err != nil {
			e.opts.errorHandler.HandleError(w, err)
			return
		}

		if rawWriter, ok := any(res).(RawWriter); ok {
			if err := rawWriter.Write(w); err != nil {
				e.opts.errorHandler.HandleError(w, err)
				return
			}

			return
		}

		statusCode := http.StatusOK

		if statusSetter, ok := any(res).(StatusSetter); ok {
			statusCode = statusSetter.StatusCode()
		}

		resBody, err := json.Marshal(res)
		if err != nil {
			e.opts.errorHandler.HandleError(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(statusCode)
		w.Write(resBody)
	})

	if len(e.opts.middleware) > 0 {
		return e.opts.middleware.Middleware(h)
	}

	return h
}
