package redis

import (
	"github.com/redis/go-redis/extra/redisotel/v9"
	"github.com/redis/go-redis/v9"
)

// InstrumentClient instruments the redis client with tracing and metrics.
func InstrumentClient(rdb redis.UniversalClient, opts ...Option) error {
	o := newOptions(opts...)

	if err := redisotel.InstrumentTracing(
		rdb,
		redisotel.WithTracerProvider(o.tp),
		redisotel.WithAttributes(o.at...),
	); err != nil {
		return err
	}

	if err := redisotel.InstrumentMetrics(
		rdb,
		redisotel.WithMeterProvider(o.mp),
		redisotel.WithAttributes(o.at...),
	); err != nil {
		return err
	}

	return nil
}
