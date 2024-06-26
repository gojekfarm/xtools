package badgerq

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/gojekfarm/xtools/xq"
)

type payload struct {
	ID string
}

func TestBadgerQueue(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	handler := func(ctx context.Context, job *xq.Job[payload]) error {
		cancel()

		return nil
	}

	q, err := New(t.TempDir(), xq.HandlerFunc[payload](handler))
	require.NoError(t, err)

	go func() {
		_ = q.Run(ctx)
	}()

	err = q.Add(ctx, &payload{ID: "1"})
	require.NoError(t, err)
}

func TestBadgerQueue_Retries(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	handler := func(ctx context.Context, job *xq.Job[payload]) error {
		if job.Attempts == 4 {
			cancel()
		}

		return assert.AnError
	}

	q, err := New(
		t.TempDir(),
		xq.HandlerFunc[payload](handler),
		MaxRetries(5),
		MaxDuration(1*time.Hour),
		Delay(10*time.Millisecond),
		Jitter(10*time.Millisecond),
		Multiplier(1.2),
	)
	require.NoError(t, err)

	go func() {
		_ = q.Run(ctx)
	}()

	err = q.Add(ctx, &payload{ID: "1"})
	require.NoError(t, err)
}

func TestBadgerQueue_DeadJobs(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	handler := func(ctx context.Context, job *xq.Job[payload]) error {
		return assert.AnError
	}

	q, err := New(
		t.TempDir(),
		xq.HandlerFunc[payload](handler),
		MaxRetries(5),
		MaxDuration(1*time.Hour),
		Delay(10*time.Millisecond),
		Jitter(10*time.Millisecond),
		Multiplier(1.2),
	)
	require.NoError(t, err)

	go func() {
		_ = q.Run(ctx)
	}()

	err = q.Add(ctx, &payload{ID: "1"})
	require.NoError(t, err)

	time.Sleep(100 * time.Millisecond)

	length, err := q.Length()
	require.NoError(t, err)
	assert.Equal(t, 1, length)

	cancel()
}
