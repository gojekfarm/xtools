package riverkfq

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"

	"github.com/gojekfarm/xtools/xkafka"
)

// PublishQueue is a thin wrapper around river.Client
// that enqueues messages are asynchronously published to Kafka.
type PublishQueue struct {
	client   *river.Client[pgx.Tx]
	pool     *pgxpool.Pool
	producer Producer

	// config
	maxWorkers int
}

// NewPublishQueue creates a new PublishQueue.
func NewPublishQueue(opts ...Option) (*PublishQueue, error) {
	pq := &PublishQueue{
		maxWorkers: 1,
	}

	for _, opt := range opts {
		opt.apply(pq)
	}

	rivercfg := &river.Config{}

	if pq.producer != nil {
		workers := river.NewWorkers()

		river.AddWorker(workers, NewPublishWorker(pq.producer))

		rivercfg.Workers = workers
		rivercfg.Queues = map[string]river.QueueConfig{
			river.QueueDefault: {MaxWorkers: pq.maxWorkers},
		}
	}

	client, err := river.NewClient(
		riverpgxv5.New(pq.pool),
		rivercfg,
	)
	if err != nil {
		return nil, err
	}

	pq.client = client

	return pq, nil
}

// Add enqueues one or more messages.
func (q *PublishQueue) Add(ctx context.Context, msgs ...*xkafka.Message) error {
	args := make([]river.InsertManyParams, len(msgs))

	for i, msg := range msgs {
		args[i] = river.InsertManyParams{
			Args: PublishArgs{Msg: msg},
		}
	}

	_, err := q.client.InsertMany(ctx, args)

	return err
}

// AddTx enqueues one or more messages in a transaction.
func (q *PublishQueue) AddTx(ctx context.Context, tx pgx.Tx, msgs ...*xkafka.Message) error {
	args := make([]river.InsertManyParams, len(msgs))

	for i, msg := range msgs {
		args[i] = river.InsertManyParams{
			Args: PublishArgs{Msg: msg},
		}
	}

	_, err := q.client.InsertManyTx(ctx, tx, args)

	return err
}

// Run starts the client and begins publishing messages to Kafka.
func (q *PublishQueue) Run(ctx context.Context) error {
	if err := q.client.Start(ctx); err != nil {
		return err
	}

	<-ctx.Done()

	if err := q.client.Stop(ctx); err != nil {
		return err
	}

	return nil
}

// PublishArgs is the job payload for publishing messages to Kafka.
type PublishArgs struct {
	Msg *xkafka.Message `json:"msg"`
}

// Kind returns the kind of the job.
func (args PublishArgs) Kind() string {
	return "publish_xkafka_message"
}

// PublishWorker is a worker that publishes messages to Kafka.
type PublishWorker struct {
	river.WorkerDefaults[PublishArgs]
	producer Producer
}

// NewPublishWorker creates a new PublishWorker.
func NewPublishWorker(producer Producer) *PublishWorker {
	return &PublishWorker{producer: producer}
}

// Work executes the job.
func (w *PublishWorker) Work(ctx context.Context, job *river.Job[PublishArgs]) error {
	return w.producer.Publish(ctx, job.Args.Msg)
}
