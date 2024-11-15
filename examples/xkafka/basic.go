package main

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"

	"github.com/gojekfarm/xrun"
	"github.com/gojekfarm/xtools/xkafka"
	logmw "github.com/gojekfarm/xtools/xkafka/middleware/zerolog"
)

func runBasic(c *cli.Context) error {
	partitions := c.Int("partitions")
	pods := c.Int("consumers")
	count := c.Int("messages")
	concurrency := c.Int("concurrency")

	ctx, cancel := context.WithCancel(c.Context)

	// create topic
	topic := createTopic(partitions)

	// generate messages
	messages := generateMessages(topic, count)

	// publish messages
	publishMessages(messages)

	tracker := NewTracker(messages, cancel)

	// consumer options
	opts := []xkafka.ConsumerOption{
		xkafka.Brokers(brokers),
		xkafka.Topics{topic},
		xkafka.ConfigMap{
			"auto.offset.reset": "earliest",
		},
		xkafka.Concurrency(concurrency),
		xkafka.ErrorHandler(func(err error) error {
			// return error to stop consumer
			return err
		}),
	}

	// create and run consumers
	runConsumers(ctx, tracker, pods, opts...)

	tracker.Summary()

	return nil
}

func basicHandler(tracker *Tracker) xkafka.HandlerFunc {
	return func(ctx context.Context, msg *xkafka.Message) error {
		err := tracker.SimulateWork()
		if err != nil {
			msg.AckFail(err)

			return err
		}

		tracker.Ack(msg)
		msg.AckSuccess()

		tracker.CancelIfDone()

		return nil
	}
}

func runConsumers(
	ctx context.Context,
	tracker *Tracker,
	pods int,
	opts ...xkafka.ConsumerOption,
) {
	log := zerolog.Ctx(ctx)
	handler := basicHandler(tracker)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			components := make([]xrun.Component, 0, pods)

			for i := 0; i < pods; i++ {
				consumer := createConsumer(handler, opts...)
				components = append(components, consumer)
			}

			err := xrun.All(xrun.NoTimeout, components...).Run(ctx)
			if err != nil {
				log.Error().Err(err).Msg("Error running consumers")
			}
		}
	}
}

func createConsumer(handler xkafka.HandlerFunc, opts ...xkafka.ConsumerOption) *xkafka.Consumer {
	consumer, err := xkafka.NewConsumer(
		"test-basic-consumer",
		handler,
		opts...,
	)
	if err != nil {
		panic(err)
	}

	consumer.Use(logmw.LoggingMiddleware(zerolog.DebugLevel))

	return consumer
}
