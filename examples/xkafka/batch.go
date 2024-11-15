package main

import (
	"context"
	"time"

	"github.com/rs/zerolog"
	"github.com/urfave/cli/v2"

	"github.com/gojekfarm/xrun"
	"github.com/gojekfarm/xtools/xkafka"
	logmw "github.com/gojekfarm/xtools/xkafka/middleware/zerolog"
)

func runBatch(c *cli.Context) error {
	partitions := c.Int("partitions")
	pods := c.Int("consumers")
	count := c.Int("messages")
	concurrency := c.Int("concurrency")
	size := c.Int("batch-size")

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
		xkafka.PollTimeout(10 * time.Second),
		xkafka.Concurrency(concurrency),
		xkafka.BatchSize(size),
		xkafka.BatchTimeout(10 * time.Second),
		xkafka.ErrorHandler(func(err error) error {
			// return error to stop consumer
			return err
		}),
	}

	// create and run consumers
	runBatchConsumers(ctx, tracker, pods, opts...)

	tracker.Summary()

	return nil
}

func batchHandler(tracker *Tracker) xkafka.BatchHandlerFunc {
	return func(ctx context.Context, batch *xkafka.Batch) error {
		err := tracker.SimulateWork()
		if err != nil {
			return batch.AckFail(err)
		}

		for _, msg := range batch.Messages {
			tracker.Ack(msg)
		}

		batch.AckSuccess()

		tracker.CancelIfDone()

		return nil
	}
}

func runBatchConsumers(ctx context.Context, tracker *Tracker, pods int, opts ...xkafka.ConsumerOption) {
	log := zerolog.Ctx(ctx)
	handler := batchHandler(tracker)

	for {
		select {
		case <-ctx.Done():
			return
		default:
			components := make([]xrun.Component, 0, pods)

			for i := 0; i < pods; i++ {
				bc, err := xkafka.NewBatchConsumer(
					"test-batch-consumer",
					handler,
					opts...,
				)
				if err != nil {
					panic(err)
				}

				bc.Use(logmw.BatchLoggingMiddleware(zerolog.DebugLevel))

				components = append(components, bc)
			}

			err := xrun.All(xrun.NoTimeout, components...).Run(ctx)
			if err != nil {
				log.Error().Err(err).Msg("Error running consumers")
			}
		}
	}
}
