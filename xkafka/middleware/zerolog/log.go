// Package zerolog provides a middleware that logs messages using zerolog.
package zerolog

import (
	"context"
	"time"

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

// BatchLoggingMiddleware is a middleware that logs batch processing using zerolog.
// Also adds a structured logger to the context.
func BatchLoggingMiddleware(lvl zerolog.Level) xkafka.BatchMiddlewareFunc {
	return func(next xkafka.BatchHandler) xkafka.BatchHandler {
		return xkafka.BatchHandlerFunc(func(ctx context.Context, b *xkafka.Batch) error {
			start := time.Now()
			fields := log.With().
				Int("count", len(b.Messages)).
				Int64("max_offset", b.MaxOffset()).
				Str("batch_id", b.ID)

			l := fields.Logger()

			ctx = l.WithContext(ctx)

			err := next.HandleBatch(ctx, b)
			if err != nil {
				l.WithLevel(zerolog.ErrorLevel).
					Dur("duration", time.Since(start)).
					Err(err).
					Msg("[xkafka] batch processing failed")
			} else {
				l.WithLevel(lvl).
					Dur("duration", time.Since(start)).
					Msg("[xkafka] batch processed")
			}

			return err
		})
	}
}
