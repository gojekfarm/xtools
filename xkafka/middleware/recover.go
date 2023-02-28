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
