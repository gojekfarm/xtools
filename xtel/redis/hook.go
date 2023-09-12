package redis

import (
	"github.com/go-redis/redis/extra/redisotel/v8"
	"github.com/go-redis/redis/v8"
)

// NewHook returns a new hook, consisting of a tracer and attribute configured
// according to the given Option(s).
func NewHook(opts ...Option) redis.Hook {
	o := newOptions(opts...)

	return redisotel.NewTracingHook(
		redisotel.WithTracerProvider(o.tp),
		redisotel.WithAttributes(o.at...),
	)
}
