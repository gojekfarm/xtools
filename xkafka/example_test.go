package xkafka

import (
	"context"
)

func ExampleConsumer() {
	handler := HandlerFunc(func(ctx context.Context, msg *Message) error {
		// do something with the message
		return nil
	})

	ignoreError := func(err error) error {
		// ignore error
		return nil
	}

	consumer, err := NewConsumer(
		"consumer-id",
		Concurrency(10),
		Topics{"test"},
		Brokers{"localhost:9092"},
		ConfigMap{
			"enable.auto.commit": false,
		},
		ErrorHandler(ignoreError),
	)
	if err != nil {
		panic(err)
	}

	consumer.
		WithHandler(handler).
		Use(
			// middleware to log messages
			MiddlewareFunc(func(next Handler) Handler {
				return HandlerFunc(func(ctx context.Context, msg *Message) error {
					// log the message
					return next.Handle(ctx, msg)
				})
			}),
		)

	if err := consumer.Start(context.Background()); err != nil {
		panic(err)
	}

	consumer.Close()
}

func ExampleProducer() {
	ctx := context.Background()

	producer, err := NewProducer(
		"producer-id",
		Brokers{"localhost:9092"},
		ConfigMap{
			"socket.keepalive.enable": true,
		},
	)
	if err != nil {
		panic(err)
	}

	producer.Use(
		// middleware to log messages
		MiddlewareFunc(func(next Handler) Handler {
			return HandlerFunc(func(ctx context.Context, msg *Message) error {
				// log the message
				return next.Handle(ctx, msg)
			})
		}),
	)

	msg := &Message{
		Topic: "test",
		Key:   []byte("key"),
		Value: []byte("value"),
	}

	if err := producer.Publish(ctx, msg); err != nil {
		panic(err)
	}

	producer.Close(ctx)
}
