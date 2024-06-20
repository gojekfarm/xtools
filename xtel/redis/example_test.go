package redis_test

import (
	"github.com/redis/go-redis/v9"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.10.0"

	xredis "github.com/gojekfarm/xtools/xtel/redis"
)

func ExampleInstrumentClient() {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	provider := sdktrace.NewTracerProvider()
	if err := xredis.InstrumentClient(rdb, xredis.WithTracerProvider(provider)); err != nil {
		panic(err)
	}
}

func ExampleWithAttributes() {
	rdb := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	if err := xredis.InstrumentClient(
		rdb,
		xredis.WithAttributes(semconv.NetPeerNameKey.String("localhost")),
	); err != nil {
		panic(err)
	}
}
