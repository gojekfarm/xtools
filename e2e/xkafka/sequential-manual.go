package main

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/gojekfarm/xrun"
	"github.com/gojekfarm/xtools/xkafka"
	"github.com/rs/xid"
	"github.com/urfave/cli/v2"
)

func runSequentialWithManualCommitTest(c *cli.Context) error {
	topic := "manual-commit-" + xid.New().String()

	if err := createTopic(topic, partitions); err != nil {
		return err
	}

	s.generated = generateMessages(topic, 10)

	// start consumers first
	// these consumers will only commit offsets for messages that are SUCCESSFULLY processed
	s.consumers = make([]*xkafka.Consumer, partitions)
	components := make([]xrun.Component, partitions)

	for i := 0; i < partitions; i++ {
		consumer, err := xkafka.NewConsumer(
			"test-seq-consumer",
			handleMessagesWithErrors(),
			xkafka.Brokers(brokers),
			xkafka.Topics{topic},
			xkafka.ManualOffset(true),
			xkafka.ConfigMap{
				"auto.offset.reset": "earliest",
			},
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

	slog.Info("[SEQUENTIAL-MANUAL] Consumers exited", "count", len(s.received))

	// create new consumers with the same group id
	// these consumers will start consuming from the last committed offset
	s.consumers = make([]*xkafka.Consumer, partitions)
	components = make([]xrun.Component, partitions)

	for i := 0; i < partitions; i++ {
		consumer, err := xkafka.NewConsumer(
			"test-seq-consumer",
			handleMessages(),
			xkafka.Brokers(brokers),
			xkafka.Topics{topic},
			xkafka.ManualOffset(true),
			xkafka.ConfigMap{
				"auto.offset.reset": "earliest",
			},
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

func handleMessagesWithErrors() xkafka.HandlerFunc {
	return func(ctx context.Context, msg *xkafka.Message) error {
		simulateWork()

		if msg.Offset > 1 {
			slog.Info("[HANDLER] Simulating error", "offset", msg.Offset)
			err := errors.New("some error")

			msg.AckFail(err)
			return err
		}

		msg.AckSuccess()

		s.received = append(s.received, msg)

		return nil
	}
}
