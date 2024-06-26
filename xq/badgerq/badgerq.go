package badgerq

import (
	"context"
	"math/rand"
	"time"

	"github.com/dgraph-io/badger/v4"
	"github.com/rs/xid"
	"github.com/vmihailenco/msgpack/v4"

	"github.com/gojekfarm/xtools/xq"
)

// BadgerQueue implements a queue using Badger.
type BadgerQueue[T any] struct {
	queue   *badger.DB
	handler xq.Handler[T]
	cfg     *config
}

// New creates a new BadgerQueue.
func New[T any](path string, handler xq.Handler[T], opts ...Option) (*BadgerQueue[T], error) {
	db, err := badger.Open(badger.DefaultOptions(path))
	if err != nil {
		return nil, err
	}

	cfg := newConfig(opts...)

	q := &BadgerQueue[T]{
		queue:   db,
		handler: handler,
		cfg:     cfg,
	}

	return q, nil
}

// Add enqueues a message to be published to Kafka.
func (q *BadgerQueue[T]) Add(ctx context.Context, msg *T) error {
	job := newJob(msg)

	return q.putJob(job)
}

// Run starts processing jobs from the queue.
func (q *BadgerQueue[T]) Run(ctx context.Context) error {
	return q.run(ctx)
}

// Length returns total number of jobs in the queue.
func (q *BadgerQueue[T]) Length() (int, error) {
	var length int

	err := q.queue.View(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = false

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			length++
		}

		return nil
	})
	if err != nil {
		return 0, err
	}

	return length, nil
}

func (q *BadgerQueue[T]) putJob(job *xq.Job[T]) error {
	return q.queue.Update(func(txn *badger.Txn) error {
		return putJobTx(txn, job)
	})
}

func (q *BadgerQueue[T]) run(ctx context.Context) error {
	for {
		select {
		case <-ctx.Done():
			return nil
		default:
			jobs, err := q.getAvailableJobs(10)
			if err != nil {
				return err
			}

			for _, job := range jobs {
				err := q.handler.Handle(ctx, job)
				if err != nil {
					job.Err = err

					return q.retryJob(job)
				}

				if err := q.completeJob(job); err != nil {
					return err
				}
			}
		}
	}
}

func (q *BadgerQueue[T]) getAvailableJobs(limit int) ([]*xq.Job[T], error) {
	var jobs []*xq.Job[T]

	err := q.queue.Update(func(txn *badger.Txn) error {
		opts := badger.DefaultIteratorOptions
		opts.PrefetchValues = true
		opts.PrefetchSize = limit
		opts.Reverse = false

		it := txn.NewIterator(opts)
		defer it.Close()

		for it.Rewind(); it.Valid(); it.Next() {
			item := it.Item()

			var job xq.Job[T]
			err := item.Value(func(val []byte) error {
				return msgpack.Unmarshal(val, &job)
			})
			if err != nil {
				return err
			}

			if job.State == xq.StateAvailable ||
				(job.State == xq.StateScheduled && job.ScheduledAt.Before(time.Now())) {
				job.Attempts++
				job.State = xq.StateInProgress
				job.ScheduledAt = nil

				if err := putJobTx(txn, &job); err != nil {
					return err
				}

				jobs = append(jobs, &job)
			}

			if len(jobs) >= limit {
				break
			}
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	return jobs, nil
}

func (q *BadgerQueue[T]) retryJob(job *xq.Job[T]) error {
	now := time.Now()

	if job.Attempts >= q.cfg.maxRetries {
		job.State = xq.StateDead
	} else if job.StartedAt != nil &&
		now.Sub(*job.StartedAt) >= q.cfg.maxDuration {
		job.State = xq.StateDead
	} else {
		job.State = xq.StateScheduled

		delay := q.cfg.delay * time.Duration(q.cfg.multiplier*float64(job.Attempts))
		delay += time.Duration(float64(q.cfg.jitter) * (1 - 2*rand.Float64()))

		scheduledAt := now.Add(delay)

		job.ScheduledAt = &scheduledAt
	}

	return q.putJob(job)
}

func (q *BadgerQueue[T]) completeJob(job *xq.Job[T]) error {
	return q.queue.Update(func(txn *badger.Txn) error {
		return txn.Delete([]byte(job.ID))
	})
}

func newJob[T any](msg *T) *xq.Job[T] {
	return &xq.Job[T]{
		ID:    xid.New().String(),
		Msg:   msg,
		State: xq.StateAvailable,
	}
}

func putJobTx[T any](txn *badger.Txn, job *xq.Job[T]) error {
	val, err := msgpack.Marshal(job)
	if err != nil {
		return err
	}

	return txn.Set([]byte(job.ID), val)
}
