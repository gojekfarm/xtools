// Package slog provides logging middleware using log/slog.
package slog

import (
	"context"
	"time"

	"log/slog"

	"github.com/gojekfarm/xtools/xkafka"
)

// Option is a configuration option for the logging middleware.
type Option interface {
	apply(*logOptions)
}

type optionFunc func(*logOptions)

func (f optionFunc) apply(o *logOptions) { f(o) }

// Level sets the log level to be used.
type Level slog.Level

func (l Level) apply(o *logOptions) { o.level = slog.Level(l) }

// Logger sets a custom logger to be used.
// slog.Default() is used by default.
func Logger(logger *slog.Logger) Option {
	return optionFunc(func(o *logOptions) {
		o.logger = logger
	})
}

type logOptions struct {
	level  slog.Level
	logger *slog.Logger
}

func newLogOptions(opts ...Option) *logOptions {
	opt := &logOptions{
		level:  slog.LevelInfo,
		logger: slog.Default(),
	}

	for _, o := range opts {
		o.apply(opt)
	}

	return opt
}

// LoggingMiddleware is a middleware that logs messages using log/slog.
func LoggingMiddleware(opts ...Option) xkafka.MiddlewareFunc {
	cfg := newLogOptions(opts...)

	return func(next xkafka.Handler) xkafka.Handler {
		return xkafka.HandlerFunc(func(ctx context.Context, msg *xkafka.Message) error {
			start := time.Now()
			logger := cfg.logger.WithGroup("xkafka")

			err := next.Handle(ctx, msg)

			args := []slog.Attr{
				slog.String("topic", msg.Topic),
				slog.Int64("partition", int64(msg.Partition)),
				slog.Int64("offset", msg.Offset),
				slog.String("key", string(msg.Key)),
				slog.String("status", msg.Status.String()),
				slog.Duration("duration", time.Since(start)),
			}

			if err != nil {
				logger.LogAttrs(ctx, slog.LevelError, "[xkafka] message processing failed", args...)
			} else {
				logger.LogAttrs(ctx, cfg.level, "[xkafka] message processed", args...)
			}

			return err
		})
	}
}
