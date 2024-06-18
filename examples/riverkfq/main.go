package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"log/slog"

	"github.com/gojekfarm/xtools/kfq/riverkfq"
	"github.com/gojekfarm/xtools/xkafka"
	slogmw "github.com/gojekfarm/xtools/xkafka/middleware/slog"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/lmittmann/tint"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivermigrate"
	"github.com/rs/xid"
)

var (
	brokers    = []string{"localhost:9092"}
	partitions = 2
)

func main() {
	slog.SetDefault(slog.New(
		tint.NewHandler(os.Stderr, &tint.Options{
			Level:      slog.LevelDebug,
			TimeFormat: time.Kitchen,
		}),
	))

	topic := "queue-" + xid.New().String()

	if err := createTopic(topic, partitions); err != nil {
		panic(err)
	}

	pool := createPool()

	pq, err := riverkfq.NewPublishQueue(pool)
	if err != nil {
		panic(err)
	}

	producer := createProducer()

	wq, err := riverkfq.NewRiverQueue(pool, producer)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithCancel(context.Background())

	go wq.Run(ctx)

	messages := generateMessages(topic, 10)

	if err := pq.Add(context.Background(), messages...); err != nil {
		panic(err)
	}

	time.Sleep(5 * time.Second)

	cancel()
	pool.Close()
	producer.Close()
}

func createProducer() *xkafka.Producer {
	producer, err := xkafka.NewProducer(
		"test-queue-producer",
		xkafka.Brokers(brokers),
		xkafka.ErrorHandler(func(err error) error {
			slog.Error("producer error", "error", err)

			return nil
		}),
	)
	if err != nil {
		panic(err)
	}

	producer.Use(
		slogmw.LoggingMiddleware(),
	)

	return producer
}

func createPool() *pgxpool.Pool {
	ctx := context.Background()

	pool, err := pgxpool.New(
		ctx,
		"postgres://postgres:postgres@localhost:5432/postgres",
	)
	if err != nil {
		panic(err)
	}

	migrator := rivermigrate.New(riverpgxv5.New(pool), nil)

	_, err = migrator.Migrate(ctx, rivermigrate.DirectionUp, &rivermigrate.MigrateOpts{})
	if err != nil {
		panic(err)
	}
	return pool
}

func generateMessages(topic string, count int) []*xkafka.Message {
	messages := make([]*xkafka.Message, count)

	for i := 0; i < count; i++ {
		messages[i] = &xkafka.Message{
			Topic: topic,
			Key:   []byte(fmt.Sprintf("key-%d", i)),
			Value: xid.New().Bytes(),
		}
	}

	return messages
}
