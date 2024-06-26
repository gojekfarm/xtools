package riverkfq

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivermigrate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/gojekfarm/xtools/xkafka"
)

func TestNewPublishQueue(t *testing.T) {
	t.Parallel()

	pool, err := setupDB(t)
	require.NoError(t, err)

	defer pool.Close()

	t.Run("OnlyEnqueue", func(t *testing.T) {
		pq, err := NewPublishQueue(
			Pool(pool),
			MaxWorkers(2),
		)
		assert.NoError(t, err)
		assert.NotNil(t, pq)
	})

	t.Run("WithProducer", func(t *testing.T) {
		pq, err := NewPublishQueue(
			Pool(pool),
			MaxWorkers(2),
			WithProducer(nil),
		)
		assert.NoError(t, err)
		assert.NotNil(t, pq)
	})
}

func TestPublishQueue_AddTx(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	pool, err := setupDB(t)
	require.NoError(t, err)

	defer pool.Close()

	pq, err := NewPublishQueue(
		Pool(pool),
		MaxWorkers(2),
	)
	require.NoError(t, err)

	msgs := []*xkafka.Message{
		{ID: "1"},
		{ID: "2"},
	}

	tx, err := pool.Begin(ctx)
	require.NoError(t, err)

	err = pq.AddTx(ctx, tx, msgs...)
	assert.Error(t, err)

	err = tx.Rollback(ctx)
	require.NoError(t, err)
}

func TestPublishQueue_E2E(t *testing.T) {
	t.Parallel()

	pool, err := setupDB(t)
	require.NoError(t, err)

	defer pool.Close()

	producer := NewMockProducer(t)

	pq, err := NewPublishQueue(
		Pool(pool),
		MaxWorkers(2),
		WithProducer(producer),
	)
	require.NoError(t, err)

	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		_ = pq.Run(ctx)
	}()

	for i := 0; i < 2; i++ {
		wg.Add(1)

		msg := &xkafka.Message{
			ID: fmt.Sprintf("%d", i),
		}

		producer.On("Publish", mock.Anything, mock.Anything).
			Return(nil).
			Once().
			Run(func(args mock.Arguments) {
				wg.Done()
			})

		err := pq.Add(ctx, msg)
		require.NoError(t, err)
	}

	wg.Wait()

	time.Sleep(1 * time.Second)
	cancel()

	producer.AssertExpectations(t)
}

func setupDB(t *testing.T) (*pgxpool.Pool, error) {
	t.Helper()

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, "postgres://postgres:postgres@localhost:5432/postgres")
	require.NoError(t, err)

	migrator := rivermigrate.New(riverpgxv5.New(pool), nil)

	_, err = migrator.Migrate(ctx, rivermigrate.DirectionUp, &rivermigrate.MigrateOpts{})

	return pool, err
}
