package xworker

// JobMiddlewareFunc functions are closures that intercept Handler calls.
type JobMiddlewareFunc func(Handler) Handler

// Middleware allows JobMiddlewareFunc to implement the handlerMiddleware interface.
func (mf JobMiddlewareFunc) Middleware(next Handler) Handler {
	return mf(next)
}

// EnqueueMiddlewareFunc functions are closures that intercept Enqueuer calls.
type EnqueueMiddlewareFunc func(Enqueuer) Enqueuer

// Middleware allows EnqueueMiddlewareFunc to implement the enqueueMiddleware interface.
func (mf EnqueueMiddlewareFunc) Middleware(next Enqueuer) Enqueuer {
	return mf(next)
}

// UseJobMiddleware appends a JobMiddlewareFunc to the chain.
// Middleware can be used to intercept or otherwise modify, process or skip Job(s).
// They are executed in the order that they are applied to the Adapter.
func (a *Adapter) UseJobMiddleware(mwf ...JobMiddlewareFunc) {
	for _, fn := range mwf {
		a.handlerMiddlewares = append(a.handlerMiddlewares, fn)
	}
}

// UseEnqueueMiddleware appends a EnqueueMiddlewareFunc to the chain.
// Middleware can be used to intercept or otherwise modify, or skip Job(s) while enqueuing.
// They are executed in the order that they are applied to the Adapter.
func (a *Adapter) UseEnqueueMiddleware(mwf ...EnqueueMiddlewareFunc) {
	for _, fn := range mwf {
		a.enqueueMiddlewares = append(a.enqueueMiddlewares, fn)
	}

	a.wrappedEnqueuer = a.enqueuer()

	for i := len(a.enqueueMiddlewares) - 1; i >= 0; i-- {
		a.wrappedEnqueuer = a.enqueueMiddlewares[i].Middleware(a.wrappedEnqueuer)
	}
}

type handlerMiddleware interface {
	// Middleware helps chain Handler(s).
	Middleware(next Handler) Handler
}

type enqueueMiddleware interface {
	// Middleware helps chain Enqueuer(s).
	Middleware(next Enqueuer) Enqueuer
}

func (a *Adapter) wrapHandlerWithMiddlewares(jobHandler Handler) Handler {
	h := jobHandler

	for i := len(a.handlerMiddlewares) - 1; i >= 0; i-- {
		h = a.handlerMiddlewares[i].Middleware(h)
	}

	return h
}
