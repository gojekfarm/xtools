package main

import (
	"context"
	"time"

	"github.com/rs/xid"
	"github.com/urfave/cli/v2"
	"log/slog"

	"github.com/gojekfarm/xrun"
	"github.com/gojekfarm/xtools/xkafka"
)

// nolint
func runAsync(c *cli.Context) error {
	topic := "async-" + xid.New().String()

	if err := createTopic(topic, partitions); err != nil {
		return err
	}

	s.generated = generateMessages(topic, 10)

	opts := []xkafka.ConsumerOption{
		xkafka.Brokers(brokers),
		xkafka.Topics{topic},
		xkafka.ConfigMap{
			"auto.offset.reset": "earliest",
		},
		xkafka.Concurrency(2),
		xkafka.ErrorHandler(func(err error) error {
			slog.Error(err.Error())
			return nil
		}),
	}

	// start consumers first
	s.consumers = make([]*xkafka.Consumer, partitions)
	components := make([]xrun.Component, partitions)

	for i := 0; i < partitions; i++ {
		consumer, err := xkafka.NewConsumer(
			"test-async-consumer",
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
			"test-async-consumer",
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

	slog.Info("[ASYNC] Messages received", "count", len(s.received))

	return nil
}
