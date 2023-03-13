package zerolog

import (
	"context"

	"github.com/gojekfarm/xtools/xkafka"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

const (
	AttributeKeyTopic     = "topic"
	AttributeKeyPartition = "partition"
	AttributeKeyMsgKey    = "msg-key"
	AttributeKeyStatus    = "status"
)

// LoggingMiddleware is a middleware that logs messages using zerolog.
func LoggingMiddleware(lvl zerolog.Level) xkafka.MiddlewareFunc {
	return func(next xkafka.Handler) xkafka.Handler {
		return xkafka.HandlerFunc(func(ctx context.Context, msg *xkafka.Message) error {
			l := log.WithLevel(lvl).
				Str(AttributeKeyTopic, msg.Topic).
				Int32(AttributeKeyPartition, msg.Partition).
				Str(AttributeKeyMsgKey, string(msg.Key))

			msg.AddCallback(func(msg *xkafka.Message) {
				l.Str(AttributeKeyStatus, msg.Status.String()).
					Msg("[XKAFKA] Message processed")
			})

			return next.Handle(ctx, msg)
		})
	}
}
