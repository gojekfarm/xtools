package xkafka

import (
	"context"
)

func ExampleConsumer() {
	handler := HandlerFunc(func(ctx context.Context, msg *Message) error {
		// do something with the message

		// acknowledge the message with success, skip or error
		msg.AckSuccess()

		return nil
	})

	ignoreError := func(err error) error {
		// ignore error
		return nil
	}

	consumer, err := NewConsumer("consumer-id", handler,
		Concurrency(10), // default is 1. values > 1 enable async processing
		Topics{"test"},
		Brokers{"localhost:9092"},
		// default behavior is to stop the consumer. this option allows customizing the error handling
		ErrorHandler(ignoreError),
		// custom configuration for the underlying kafka consumer
		ConfigMap{
			"auto.commit.interval.ms": 1000,
		},
		// default behavior is to commit messages automatically.
		// this option triggers manual commit after each message is processed.
		ManualCommit(true),
	)
	if err != nil {
		panic(err)
	}

	consumer.Use(
		// middleware to log messages
		MiddlewareFunc(func(next Handler) Handler {
			return HandlerFunc(func(ctx context.Context, msg *Message) error {
				// log the message
				return next.Handle(ctx, msg)
			})
		}),
	)

	if err := consumer.Run(context.Background()); err != nil {
		panic(err)
	}

	consumer.Close()
}

func ExampleProducer() {
	ctx, cancel := context.WithCancel(context.Background())

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

	go func() {
		err := producer.Run(ctx)
		if err != nil {
			panic(err)
		}
	}()

	msg := &Message{
		Topic: "test",
		Key:   []byte("key"),
		Value: []byte("value"),
	}

	if err := producer.Publish(ctx, msg); err != nil {
		panic(err)
	}

	// cancel the context to stop the producer
	cancel()
}

func ExampleProducer_AsyncPublish() {
	ctx, cancel := context.WithCancel(context.Background())

	// default callback function called after each message
	// handled by the producer
	callback := func(msg *Message) {
		// do something with the message
	}

	producer, err := NewProducer(
		"producer-id",
		Brokers{"localhost:9092"},
		ConfigMap{
			"socket.keepalive.enable": true,
		},
		DeliveryCallback(callback),
	)
	if err != nil {
		panic(err)
	}

	go func() {
		err := producer.Run(ctx)
		if err != nil {
			panic(err)
		}
	}()

	msg := &Message{
		Topic: "test",
		Key:   []byte("key"),
		Value: []byte("value"),
	}

	// each message can have its own callback function
	// in addition to the default callback function
	msg.AddCallback(func(m *Message) {
		// do something with the message
	})

	err = producer.AsyncPublish(ctx, msg)
	if err != nil {
		panic(err)
	}

	// cancel the context to stop the producer
	cancel()
}
