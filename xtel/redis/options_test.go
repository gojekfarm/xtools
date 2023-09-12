package redis

import (
	"context"
	"testing"

	"github.com/go-redis/redis/v8"
	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/attribute"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"
	"go.opentelemetry.io/otel/trace"
)

func TestWithAttributes(t *testing.T) {
	provider := sdktrace.NewTracerProvider()
	hook := NewHook(
		WithTracerProvider(provider),
		WithAttributes(semconv.NetPeerNameKey.String("localhost")))

	ctx, span := provider.Tracer("redis-test").Start(context.TODO(), "redis-test")
	cmd := redis.NewCmd(ctx, "ping")
	defer span.End()

	ctx, err := hook.BeforeProcess(ctx, cmd)
	assert.NoError(t, err)

	assert.NoError(t, hook.AfterProcess(ctx, cmd))

	attrs := trace.SpanFromContext(ctx).(sdktrace.ReadOnlySpan).Attributes()
	assert.ElementsMatch(t, []attribute.KeyValue{
		semconv.DBSystemRedis,
		semconv.NetPeerNameKey.String("localhost"),
		semconv.DBStatementKey.String("ping"),
	}, attrs)
}
