package retry_test

import (
	"context"
	"log/slog"
	"time"

	"github.com/gojekfarm/xtools/xkafka"
	"github.com/gojekfarm/xtools/xkafka/middleware/retry"
)

func Example() {
	handler := func(ctx context.Context, m *xkafka.Message) error {
		// handle message
		return nil
	}

	consumer, err := xkafka.NewConsumer(
		"retry-consumer",
		xkafka.HandlerFunc(handler),
		xkafka.Brokers{"localhost:9092"},
		xkafka.Topics{"test-topic"},
		xkafka.ErrorHandler(func(err error) error {
			slog.Error(err.Error())

			return nil
		}),
	)
	if err != nil {
		panic(err)
	}

	consumer.Use(
		retry.ExponentialBackoff(
			retry.MaxRetries(3),
			retry.MaxLifetime(10*time.Second),
			retry.Delay(1*time.Second),
			retry.Jitter(100*time.Millisecond),
		),
	)

	// ... run consumer
}
