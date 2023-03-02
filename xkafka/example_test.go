package xkafka_test

import (
	"context"

	"github.com/gojekfarm/xtools/xkafka"
)

func ExampleConsumer() {
	handler := xkafka.HandlerFunc(func(ctx context.Context, msg *xkafka.Message) error {
		// do something with the message
		return nil
	})

	ignoreError := func(err error) error {
		// ignore error
		return nil
	}

	consumer, err := xkafka.NewConsumer(
		"consumer-id",
		xkafka.Concurrency(10),
		xkafka.Topics([]string{"test"}),
		xkafka.Brokers([]string{"localhost:9092"}),
		xkafka.ConfigMap(xkafka.ConfigMap{
			"enable.auto.commit": false,
		}),
		xkafka.ErrorHandler(ignoreError),
	)
	if err != nil {
		panic(err)
	}

	consumer.Use(
		// middleware to log messages
		xkafka.MiddlewareFunc(func(next xkafka.Handler) xkafka.Handler {
			return xkafka.HandlerFunc(func(ctx context.Context, msg *xkafka.Message) error {
				// log the message
				return next.Handle(ctx, msg)
			})
		}),
	)

	if err := consumer.Start(context.Background(), handler); err != nil {
		panic(err)
	}

	consumer.Close()
}
