// Package zerolog provides a middleware that logs messages using zerolog.
package zerolog

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/gojekfarm/xtools/xkafka"
)

// Log attribute names.
const (
	AttributeKeyTopic     = "topic"
	AttributeKeyPartition = "partition"
	AttributeKeyMsgKey    = "msg-key"
	AttributeKeyStatus    = "status"
)

// LoggingMiddleware is a middleware that logs messages using zerolog.
// Also adds a structured logger to the context.
func LoggingMiddleware(lvl zerolog.Level) xkafka.MiddlewareFunc {
	return func(next xkafka.Handler) xkafka.Handler {
		return xkafka.HandlerFunc(func(ctx context.Context, msg *xkafka.Message) error {
			fields := log.With().
				Str(AttributeKeyTopic, msg.Topic).
				Int32(AttributeKeyPartition, msg.Partition).
				Str(AttributeKeyMsgKey, string(msg.Key))

			l := fields.Logger()

			msg.AddCallback(func(msg *xkafka.Message) {
				l.
					WithLevel(lvl).
					Str(AttributeKeyStatus, msg.Status.String()).
					Msg("[xkafka] message processed")
			})

			ctx = l.WithContext(ctx)

			return next.Handle(ctx, msg)
		})
	}
}
