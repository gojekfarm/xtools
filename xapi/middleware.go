package xapi

import "net/http"

// MiddlewareHandler defines the interface for HTTP middleware.
type MiddlewareHandler interface {
	Middleware(next http.Handler) http.Handler
}

// MiddlewareFunc is a function type that implements MiddlewareHandler.
type MiddlewareFunc func(next http.Handler) http.Handler

// Middleware implements the MiddlewareHandler interface.
func (m MiddlewareFunc) Middleware(next http.Handler) http.Handler {
	return m(next)
}

// MiddlewareStack represents a stack of middleware handlers.
type MiddlewareStack []MiddlewareHandler

// Middleware applies all middleware in the stack to the given handler.
// Middleware is applied in reverse order, so the last added middleware
// wraps the innermost handler.
func (m MiddlewareStack) Middleware(next http.Handler) http.Handler {
	for i := len(m) - 1; i >= 0; i-- {
		next = m[i].Middleware(next)
	}
	return next
}
