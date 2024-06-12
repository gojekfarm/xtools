package retry_test

import (
	"context"
	"time"

	"log/slog"

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
			retry.MaxRetries(3),                // retry 3 times
			retry.MaxDuration(10*time.Second),  // don't retry after 10 seconds
			retry.Delay(1*time.Second),         // initial delay
			retry.Jitter(100*time.Millisecond), // random delay to avoid thundering herd
			retry.Multiplier(1.5),              // multiplier for exponential backoff
		),
		// add other middlewares after retry to run them
		// on each retry, like logging, metrics, etc.
	)

	// ... run consumer
}
