package xapi

type options struct {
	middleware   MiddlewareStack
	errorHandler ErrorHandler
}

// EndpointOption defines the interface for endpoint configuration options.
type EndpointOption interface {
	apply(o *options)
}

// endpointOptionFunc is a function type that implements EndpointOption.
type endpointOptionFunc func(o *options)

// apply implements the EndpointOption interface.
func (f endpointOptionFunc) apply(o *options) {
	f(o)
}

// WithMiddleware returns an EndpointOption that adds middleware to the endpoint.
func WithMiddleware(middlewares ...MiddlewareHandler) EndpointOption {
	return endpointOptionFunc(func(o *options) {
		o.middleware = append(o.middleware, middlewares...)
	})
}

// WithErrorHandler returns an EndpointOption that sets a custom error handler for the endpoint.
func WithErrorHandler(errorHandler ErrorHandler) EndpointOption {
	return endpointOptionFunc(func(o *options) {
		o.errorHandler = errorHandler
	})
}
