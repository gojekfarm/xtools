package main

import (
	"context"

	"log/slog"

	"github.com/gojekfarm/xtools/xkafka"
)

func publishMessages(messages []*xkafka.Message) {
	producer, err := xkafka.NewProducer(
		"test-seq-producer",
		xkafka.Brokers(brokers),
		xkafka.ErrorHandler(func(err error) error {
			slog.Error(err.Error())

			return err
		}),
	)
	if err != nil {
		panic(err)
	}

	defer producer.Close()

	for _, msg := range messages {
		if err := producer.Publish(context.Background(), msg); err != nil {
			panic(err)
		}
	}

	slog.Info("[PRODUCER] published messages", "count", len(messages))
}
