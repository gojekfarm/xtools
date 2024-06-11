package retry

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/gojekfarm/xtools/xkafka"
)

func TestExponentialBackoff_MaxRetries(t *testing.T) {
	msg := &xkafka.Message{
		Topic:     "test-topic",
		Group:     "test-group",
		Partition: 2,
		Key:       []byte("key"),
		Value:     []byte("value"),
	}

	attempts := 0

	handler := xkafka.HandlerFunc(func(ctx context.Context, m *xkafka.Message) error {
		attempts++

		return assert.AnError
	})

	mw := ExponentialBackoff(
		MaxRetries(3),
	)

	err := mw(handler).Handle(context.TODO(), msg)
	assert.ErrorIs(t, err, assert.AnError)
	assert.Equal(t, 4, attempts)
}

func TestExponentialBackoff_MaxLifetime(t *testing.T) {
	msg := &xkafka.Message{
		Topic:     "test-topic",
		Group:     "test-group",
		Partition: 2,
		Key:       []byte("key"),
		Value:     []byte("value"),
	}

	handler := xkafka.HandlerFunc(func(ctx context.Context, m *xkafka.Message) error {
		return assert.AnError
	})

	mw := ExponentialBackoff(
		MaxRetries(1000),
		MaxLifetime(1*time.Second),
		Delay(10*time.Millisecond),
		Jitter(2*time.Millisecond),
		Multiplier(1.5),
	)

	start := time.Now()
	err := mw(handler).Handle(context.TODO(), msg)
	assert.ErrorIs(t, err, assert.AnError)
	assert.WithinDuration(t, start, time.Now(), 1*time.Second)
}

func TestExponentialBackoff_PermanentError(t *testing.T) {
	msg := &xkafka.Message{
		Topic:     "test-topic",
		Group:     "test-group",
		Partition: 2,
		Key:       []byte("key"),
		Value:     []byte("value"),
	}

	handler := xkafka.HandlerFunc(func(ctx context.Context, m *xkafka.Message) error {
		return ErrPermanent
	})

	mw := ExponentialBackoff(MaxRetries(1000))

	err := mw(handler).Handle(context.TODO(), msg)
	assert.ErrorIs(t, err, ErrPermanent)
}
