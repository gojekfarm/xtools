package slog

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"log/slog"

	"github.com/gojekfarm/xtools/xkafka"
)

var logger = Logger(slog.Default())

func TestLoggingMiddleware(t *testing.T) {
	msg := &xkafka.Message{
		Topic:     "test-topic",
		Partition: 0,
		Offset:    0,
		Key:       []byte("test-key"),
	}

	loggingMiddleware := LoggingMiddleware(
		Level(slog.LevelInfo),
		logger,
	)
	handler := loggingMiddleware(xkafka.HandlerFunc(func(ctx context.Context, msg *xkafka.Message) error {
		msg.AckSuccess()

		return nil
	}))

	err := handler.Handle(context.Background(), msg)
	assert.NoError(t, err)
}

func TestLoggingMiddlewareWithError(t *testing.T) {
	msg := &xkafka.Message{
		Topic:     "test-topic",
		Partition: 0,
		Offset:    0,
		Key:       []byte("test-key"),
	}

	loggingMiddleware := LoggingMiddleware(
		Level(slog.LevelInfo),
		logger,
	)
	handler := loggingMiddleware(xkafka.HandlerFunc(func(ctx context.Context, msg *xkafka.Message) error {
		msg.AckFail(assert.AnError)

		return assert.AnError
	}))

	err := handler.Handle(context.Background(), msg)
	assert.ErrorIs(t, err, assert.AnError)
}
