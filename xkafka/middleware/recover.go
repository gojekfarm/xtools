package middleware

import (
	"context"
	"fmt"

	"github.com/gojekfarm/xtools/xkafka"
)

// RecoverMiddleware catches panics and prevents the request goroutine from crashing.
func RecoverMiddleware() xkafka.MiddlewareFunc {
	return func(next xkafka.Handler) xkafka.Handler {
		return xkafka.HandlerFunc(func(ctx context.Context, m *xkafka.Message) error {
			defer func() {
				if r := recover(); r != nil {
					m.AckFail(fmt.Errorf("%+v", r))

					return
				}
			}()

			return next.Handle(ctx, m)
		})
	}
}

// BatchRecoverMiddleware catches panics and prevents the request goroutine from crashing
// for xkafka.BatchConsumer.
func BatchRecoverMiddleware() xkafka.BatchMiddlewareFunc {
	return func(next xkafka.BatchHandler) xkafka.BatchHandler {
		return xkafka.BatchHandlerFunc(func(ctx context.Context, batch *xkafka.Batch) error {
			defer func() {
				if r := recover(); r != nil {
					_ = batch.AckFail(fmt.Errorf("%+v", r))

					return
				}
			}()

			return next.HandleBatch(ctx, batch)
		})
	}
}
