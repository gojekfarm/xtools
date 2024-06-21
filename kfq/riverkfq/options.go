package riverkfq

import (
	"context"

	"github.com/gojekfarm/xtools/xkafka"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
)

// Option defines interface for configuring riverkfq
type Option interface {
	apply(*PublishQueue)
}

type optionFunc func(*PublishQueue)

func (f optionFunc) apply(pq *PublishQueue) {
	f(pq)
}

// Pool sets the database connection pool.
func Pool(p *pgxpool.Pool) Option {
	return optionFunc(func(pq *PublishQueue) {
		pq.pool = p
	})
}

// Producer is a Kafka producer
type Producer interface {
	Publish(ctx context.Context, msg *xkafka.Message) error
}

// WithProducer sets the Kafka producer.
func WithProducer(p Producer) Option {
	return optionFunc(func(pq *PublishQueue) {
		pq.producer = p
	})
}

// WithClient sets the river.Client.
// Useful for integrating with existing river queues.
func WithClient(c *river.Client[pgx.Tx]) Option {
	return optionFunc(func(pq *PublishQueue) {
		pq.queue = c
	})
}

// MaxWorkers sets the maximum number of workers.
// Default is 1.
type MaxWorkers int

func (m MaxWorkers) apply(pq *PublishQueue) {
	pq.maxWorkers = int(m)
}
