package xkafka

import (
	"context"
)

// Handler defines a message handler.
type Handler interface {
	Handle(ctx context.Context, m *Message) error
}

// HandlerFunc defines a function for handling messages.
type HandlerFunc func(ctx context.Context, m *Message) error

// Handle implements Handler interface on HandlerFunc.
func (h HandlerFunc) Handle(ctx context.Context, m *Message) error {
	return h(ctx, m)
}

// Middlewarer is an interface for message handler middleware.
type Middlewarer interface {
	Middleware(handler Handler) Handler
}

// MiddlewareFunc defines a function for middleware.
type MiddlewareFunc func(Handler) Handler

// Middleware implements Middlewarer interface.
func (mw MiddlewareFunc) Middleware(handler Handler) Handler {
	return mw(handler)
}

// BatchHandler defines a handler for a batch of messages.
type BatchHandler interface {
	HandleBatch(ctx context.Context, b *Batch) error
}

// BatchHandlerFunc defines a function for handling a batch.
type BatchHandlerFunc func(ctx context.Context, b *Batch) error

// HandleBatch implements BatchHandler interface.
func (h BatchHandlerFunc) HandleBatch(ctx context.Context, b *Batch) error {
	return h(ctx, b)
}

// BatchMiddlewarer is an interface for batch message handler middleware.
type BatchMiddlewarer interface {
	BatchMiddleware(handler BatchHandler) BatchHandler
}

// BatchMiddlewareFunc defines a function for batch middleware.
type BatchMiddlewareFunc func(BatchHandler) BatchHandler

// BatchMiddleware implements BatchMiddlewarer interface.
func (mw BatchMiddlewareFunc) BatchMiddleware(handler BatchHandler) BatchHandler {
	return mw(handler)
}
