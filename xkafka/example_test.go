package xkafka_test

import (
	"context"

	"github.com/gojekfarm/xtools/xkafka"
)

func ExampleSimpleConsumer() {
	cfg := xkafka.ConsumerConfig{
		Topics:  []string{"test"},
		Brokers: "localhost:9092",
		Group:   "test",
	}

	handler := xkafka.HandlerFunc(func(ctx context.Context, msg *xkafka.Message) error {
		// do something with the message
		return nil
	})

	ignoreError := func(err error) error {
		// ignore error
		return nil
	}

	consumer, err := xkafka.NewSimpleConsumer(cfg)
	if err != nil {
		panic(err)
	}

	consumer.SetErrorHandler(ignoreError)

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
