package riverkfq

import (
	"context"

	"github.com/gojekfarm/xtools/xkafka"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
)

// Producer is a Kafka producer
type Producer interface {
	Publish(ctx context.Context, msg *xkafka.Message) error
}

// RiverQueue guarantees that messages are published
// to Kafka.
type RiverQueue struct {
	queue *river.Client[pgx.Tx]
}

// NewPublishQueue creates a new RiverQueue which can only
// queue messages.
func NewPublishQueue(pool *pgxpool.Pool) (*RiverQueue, error) {
	client, err := river.NewClient(
		riverpgxv5.New(pool),
		&river.Config{},
	)
	if err != nil {
		return nil, err
	}

	return &RiverQueue{
		queue: client,
	}, nil
}

// NewRiverQueue creates a new RiverQueue which can publish
// messages to Kafka.
func NewRiverQueue(pool *pgxpool.Pool, producer Producer) (*RiverQueue, error) {
	workers := river.NewWorkers()
	river.AddWorker(workers, &PublishWorker{producer: producer})

	client, err := river.NewClient(riverpgxv5.New(pool), &river.Config{
		Queues: map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: 3},
		},
		Workers: workers,
	})
	if err != nil {
		return nil, err
	}

	return &RiverQueue{
		queue: client,
	}, nil
}

// Add enqueues one or more messages.
func (q *RiverQueue) Add(ctx context.Context, msgs ...*xkafka.Message) error {
	args := make([]river.InsertManyParams, len(msgs))

	for i, msg := range msgs {
		args[i] = river.InsertManyParams{
			Args: PublishArgs{Msg: msg},
		}
	}

	_, err := q.queue.InsertMany(ctx, args)

	return err
}

// AddTx enqueues one or more messages in a transaction.
func (q *RiverQueue) AddTx(ctx context.Context, tx pgx.Tx, msgs ...*xkafka.Message) error {
	args := make([]river.InsertManyParams, len(msgs))

	for i, msg := range msgs {
		args[i] = river.InsertManyParams{
			Args: PublishArgs{Msg: msg},
		}
	}

	_, err := q.queue.InsertManyTx(ctx, tx, args)

	return err
}

// Run starts the queue and begins publishing messages to Kafka.
func (q *RiverQueue) Run(ctx context.Context) error {
	if err := q.queue.Start(ctx); err != nil {
		return err
	}

	<-ctx.Done()

	if err := q.queue.Stop(ctx); err != nil {
		return err
	}

	return nil
}

type PublishArgs struct {
	Msg *xkafka.Message `json:"msg"`
}

func (args PublishArgs) Kind() string {
	return "publish_xkafka_message"
}

type PublishWorker struct {
	river.WorkerDefaults[PublishArgs]
	producer Producer
}

func NewPublishWorker(producer Producer) *PublishWorker {
	return &PublishWorker{
		producer: producer,
	}
}

func (w *PublishWorker) Work(ctx context.Context, job *river.Job[PublishArgs]) error {
	return w.producer.Publish(ctx, job.Args.Msg)
}
