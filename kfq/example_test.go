package kfq

import (
	"context"
	"sync"
	"time"

	"github.com/gojekfarm/xtools/xkafka"
)

func ExampleQueue_producer() {
	ctx, cancel := context.WithCancel(context.Background())

	// Create a new queue with default options.
	q := NewQueue("producer-queue")

	// Create a new queue with custom options.
	q = NewQueue(
		"producer-queue",
		MaxRetries(10),
		MaxDuration(1*time.Minute),
	)

	// Create a new message.
	msg := &xkafka.Message{
		Topic: "test-topic",
		Key:   []byte("test-key"),
		Value: []byte("test-value"),
	}

	producer, err := xkafka.NewProducer("localhost:9092")
	if err != nil {
		panic(err)
	}

	producer.Use(
		q.Middleware(),
	)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		if err := producer.Run(ctx); err != nil {
			panic(err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		if err := q.Run(ctx); err != nil {
			panic(err)
		}
	}()

	// Publish a message
	if err := producer.Publish(ctx, msg); err != nil {
		panic(err)
	}

	cancel()
}

func ExampleQueue_consumer() {
	ctx, cancel := context.WithCancel(context.Background())

	// Create a new queue with default options.
	q := NewQueue("consumer-queue")

	// Create a new queue with custom options.
	q = NewQueue(
		"consumer-queue",
		MaxRetries(10),
		MaxDuration(1*time.Minute),
	)

	handler := func(ctx context.Context, msg *xkafka.Message) error {
		return nil
	}

	consumer, err := xkafka.NewConsumer("localhost:9092", xkafka.HandlerFunc(handler))
	if err != nil {
		panic(err)
	}

	consumer.Use(
		q.Middleware(),
	)

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		if err := consumer.Run(ctx); err != nil {
			panic(err)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		if err := q.Run(ctx); err != nil {
			panic(err)
		}
	}()

	cancel()
}
