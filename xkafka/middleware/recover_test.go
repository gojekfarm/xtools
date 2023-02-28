package middleware_test

import (
	"context"
	"testing"

	"github.com/gojekfarm/xtools/xkafka"
	"github.com/gojekfarm/xtools/xkafka/middleware"
	"github.com/stretchr/testify/assert"
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
