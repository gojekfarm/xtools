package zerolog

import (
	"context"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"

	"github.com/gojekfarm/xtools/xkafka"
)

func TestLoggingMiddleware(t *testing.T) {
	msg := &xkafka.Message{
		Topic:     "test-topic",
		Partition: 0,
		Offset:    0,
		Key:       []byte("test-key"),
	}

	loggingMiddleware := LoggingMiddleware(zerolog.InfoLevel)
	t.Run("success", func(t *testing.T) {
		handler := loggingMiddleware(xkafka.HandlerFunc(func(ctx context.Context, msg *xkafka.Message) error {
			msg.AckSuccess()
			return nil
		}))

		err := handler.Handle(context.Background(), msg)
		assert.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		handler := loggingMiddleware(xkafka.HandlerFunc(func(ctx context.Context, msg *xkafka.Message) error {
			msg.AckFail(assert.AnError)
			return assert.AnError
		}))

		err := handler.Handle(context.Background(), msg)
		assert.ErrorIs(t, err, assert.AnError)
	})
}

func TestBatchLoggingMiddleware(t *testing.T) {
	batch := xkafka.NewBatch()
	msg := &xkafka.Message{
		Topic:     "test-topic",
		Partition: 0,
		Offset:    0,
		Key:       []byte("test-key"),
	}

	batch.Messages = append(batch.Messages, msg)

	loggingMiddleware := BatchLoggingMiddleware(zerolog.InfoLevel)

	t.Run("success", func(t *testing.T) {
		handler := loggingMiddleware(xkafka.BatchHandlerFunc(func(ctx context.Context, b *xkafka.Batch) error {
			b.AckSuccess()
			return nil
		}))

		err := handler.HandleBatch(context.Background(), batch)
		assert.NoError(t, err)
	})

	t.Run("error", func(t *testing.T) {
		handler := loggingMiddleware(xkafka.BatchHandlerFunc(func(ctx context.Context, b *xkafka.Batch) error {
			b.AckFail(assert.AnError)
			return assert.AnError
		}))

		err := handler.HandleBatch(context.Background(), batch)
		assert.ErrorIs(t, err, assert.AnError)
	})
}
