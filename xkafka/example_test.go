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

func ExampleProducer() {
	ctx := context.Background()

	producer, err := xkafka.NewProducer(
		xkafka.Brokers([]string{"localhost:9092"}),
		xkafka.ConfigMap(xkafka.ConfigMap{
			"socket.keepalive.enable": true,
			"client.id":               "test-producer",
		}),
	)
	if err != nil {
		panic(err)
	}

	producer.Use(
		// middleware to log messages
		xkafka.MiddlewareFunc(func(next xkafka.Handler) xkafka.Handler {
			return xkafka.HandlerFunc(func(ctx context.Context, msg *xkafka.Message) error {
				// log the message
				return next.Handle(ctx, msg)
			})
		}),
	)

	msg := &xkafka.Message{
		Topic: "test",
		Key:   []byte("key"),
		Value: []byte("value"),
	}

	if err := producer.Publish(ctx, msg); err != nil {
		panic(err)
	}

	producer.Close(ctx)
}
