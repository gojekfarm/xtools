package riverkfq

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/riverqueue/river/riverdriver/riverpgxv5"
	"github.com/riverqueue/river/rivermigrate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gojekfarm/xtools/xkafka"
)

func TestNewPublishQueue(t *testing.T) {
	t.Run("OnlyEnqueue", func(t *testing.T) {
		pq, err := NewPublishQueue(
			Pool(nil),
			MaxWorkers(2),
		)
		assert.NoError(t, err)
		assert.NotNil(t, pq)
	})

	t.Run("WithProducer", func(t *testing.T) {
		pq, err := NewPublishQueue(
			Pool(nil),
			MaxWorkers(2),
			WithProducer(nil),
		)
		assert.NoError(t, err)
		assert.NotNil(t, pq)
	})
}

func TestPublishQueue_Add(t *testing.T) {
	ctx := context.Background()

	pool, err := setupDB(t)
	require.NoError(t, err)

	defer pool.Close()

	pq, err := NewPublishQueue(
		Pool(pool),
	)
	require.NoError(t, err)

	msgs := []*xkafka.Message{
		{ID: "1"},
		{ID: "2"},
	}

	t.Run("Add", func(t *testing.T) {
		err := pq.Add(ctx, msgs...)
		assert.NoError(t, err)
	})

	t.Run("AddTx", func(t *testing.T) {
		tx, err := pool.Begin(ctx)
		require.NoError(t, err)

		err = pq.AddTx(ctx, tx, msgs...)
		assert.NoError(t, err)

		err = tx.Commit(ctx)
		require.NoError(t, err)
	})
}

func setupDB(t *testing.T) (*pgxpool.Pool, error) {
	t.Helper()

	createTestDB(t)

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, "postgres://postgres:postgres@localhost:5432/riverkfq_test")
	require.NoError(t, err)

	migrator := rivermigrate.New(riverpgxv5.New(pool), nil)

	_, err = migrator.Migrate(ctx, rivermigrate.DirectionUp, &rivermigrate.MigrateOpts{})

	return pool, err
}

func createTestDB(t *testing.T) {
	t.Helper()

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, "postgres://postgres:postgres@localhost:5432/postgres")
	require.NoError(t, err)

	_, err = pool.Exec(ctx, "CREATE DATABASE riverkfq_test")
	require.NoError(t, err)

	t.Cleanup(func() {
		_, err = pool.Exec(ctx, "DROP DATABASE riverkfq_test")
		require.NoError(t, err)

		pool.Close()
	})
}
