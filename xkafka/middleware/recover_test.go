package middleware_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gojekfarm/xtools/xkafka"
	"github.com/gojekfarm/xtools/xkafka/middleware"
)

func TestRecoverMiddleware(t *testing.T) {
	var handler xkafka.Handler

	handler = xkafka.HandlerFunc(func(ctx context.Context, msg *xkafka.Message) error {
		panic("test panic")
	})

	m := middleware.RecoverMiddleware()

	handler = m.Middleware(handler)
	msg := xkafka.Message{}

	err := handler.Handle(context.Background(), &msg)
	assert.NoError(t, err)
	assert.Equal(t, xkafka.Fail, msg.Status)
}

func TestBatchRecoverMiddleware(t *testing.T) {
	var handler xkafka.BatchHandler

	handler = xkafka.BatchHandlerFunc(func(ctx context.Context, batch *xkafka.Batch) error {
		panic("test panic")
	})

	m := middleware.BatchRecoverMiddleware()

	handler = m.BatchMiddleware(handler)
	batch := xkafka.Batch{}

	err := handler.HandleBatch(context.Background(), &batch)
	assert.NoError(t, err)
	assert.Equal(t, xkafka.Fail, batch.Status)
}
