package redis_test

import (
	"context"

	"github.com/go-redis/redis/v8"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"

	xredis "github.com/gojekfarm/xtools/xtel/redis"
)

func ExampleNewHook() {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	h := xredis.NewHook()
	rdb.AddHook(h)
}

func ExampleWithAttributes() {
	provider := sdktrace.NewTracerProvider()
	xredis.NewHook(
		xredis.WithTracerProvider(provider),
		xredis.WithAttributes(semconv.NetPeerNameKey.String("localhost")))

	_, span := provider.Tracer("redis-example").Start(context.TODO(), "redis-example")
	defer span.End()
}
