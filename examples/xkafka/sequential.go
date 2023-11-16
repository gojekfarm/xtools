package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/rs/xid"
	"github.com/urfave/cli/v2"
	"log/slog"

	"github.com/gojekfarm/xrun"
	"github.com/gojekfarm/xtools/xkafka"
)

var (
	brokers    = []string{"localhost:9092"}
	partitions = 2
)

// nolint
func runSequential(c *cli.Context) error {
	topic := "seq-" + xid.New().String()

	if err := createTopic(topic, partitions); err != nil {
		return err
	}

	s.generated = generateMessages(topic, 10)

	opts := []xkafka.Option{
		xkafka.Brokers(brokers),
		xkafka.Topics{topic},
		xkafka.ConfigMap{
			"auto.offset.reset": "earliest",
		},
	}

	// start consumers first
	s.consumers = make([]*xkafka.Consumer, partitions)
	components := make([]xrun.Component, partitions)

	for i := 0; i < partitions; i++ {
		consumer, err := xkafka.NewConsumer(
			"test-seq-consumer",
			handleMessagesWithErrors(),
			opts...,
		)
		if err != nil {
			panic(err)
		}

		consumer.Use(loggingMiddleware())

		s.consumers[i] = consumer
		components[i] = consumer
	}

	// publish messages to the topic
	if err := publishMessages(s.generated); err != nil {
		return err
	}

	runConsumers(c.Context, s.consumers)

	slog.Info("[SEQUENTIAL] Consumers exited", "count", len(s.received))

	// create new consumers with the same group id
	// these consumers will start consuming from the last committed offset
	s.consumers = make([]*xkafka.Consumer, partitions)
	components = make([]xrun.Component, partitions)

	for i := 0; i < partitions; i++ {
		consumer, err := xkafka.NewConsumer(
			"test-seq-consumer",
			handleMessages(),
			opts...,
		)
		if err != nil {
			panic(err)
		}

		consumer.Use(loggingMiddleware())

		s.consumers[i] = consumer
		components[i] = consumer
	}

	ctx, cancel := context.WithCancel(c.Context)

	go runConsumers(ctx, s.consumers)

	// wait for all messages to be processed
	for {
		if len(s.received) == len(s.generated) {
			<-time.After(1 * time.Second)

			cancel()

			break
		}
	}

	return nil
}

func publishMessages(messages []*xkafka.Message) error {
	producer, err := xkafka.NewProducer(
		"test-seq-producer",
		xkafka.Brokers(brokers),
	)
	if err != nil {
		return err
	}

	defer producer.Close()

	for _, msg := range messages {
		if err := producer.Publish(context.Background(), msg); err != nil {
			return err
		}

		slog.Info("[PRODUCER] published message", "key", string(msg.Key))
	}

	return nil
}

func generateMessages(topic string, count int) []*xkafka.Message {
	messages := make([]*xkafka.Message, count)

	for i := 0; i < count; i++ {
		messages[i] = &xkafka.Message{
			Topic: topic,
			Key:   []byte(fmt.Sprintf("key-%d", i)),
			Value: xid.New().Bytes(),
		}
	}

	return messages
}

func loggingMiddleware() xkafka.MiddlewareFunc {
	return func(next xkafka.Handler) xkafka.Handler {
		return xkafka.HandlerFunc(func(ctx context.Context, msg *xkafka.Message) error {
			start := time.Now()

			err := next.Handle(ctx, msg)

			slog.Info("[MESSAGE] Completed",
				"key", string(msg.Key),
				"offset", msg.Offset, "partition", msg.Partition,
				"duration", time.Since(start),
				"status", msg.Status,
			)

			return err
		})
	}
}

func simulateWork() {
	<-time.After(time.Duration(rand.Int63n(200)) * time.Millisecond)
}

func runConsumers(ctx context.Context, consumers []*xkafka.Consumer) {
	components := make([]xrun.Component, len(consumers))

	for i, consumer := range consumers {
		components[i] = consumer
	}

	err := xrun.All(xrun.NoTimeout, components...).Run(ctx)
	if err != nil {
		panic(err)
	}
}

func handleMessages() xkafka.HandlerFunc {
	return func(ctx context.Context, msg *xkafka.Message) error {
		s.mu.Lock()
		s.received[string(msg.Key)] = msg
		s.mu.Unlock()

		simulateWork()

		msg.AckSuccess()

		return nil
	}
}

func handleMessagesWithErrors() xkafka.HandlerFunc {
	return func(ctx context.Context, msg *xkafka.Message) error {
		s.mu.Lock()
		s.received[string(msg.Key)] = msg
		s.mu.Unlock()

		simulateWork()

		if msg.Offset > 1 {
			err := fmt.Errorf("simulated error for key %s", string(msg.Key))

			msg.AckFail(err)

			return err
		}

		msg.AckSuccess()

		return nil
	}
}
