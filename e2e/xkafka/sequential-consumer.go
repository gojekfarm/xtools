package main

import (
	"context"
	"fmt"

	"log/slog"

	"github.com/gojekfarm/xtools/xkafka"
	"github.com/rs/xid"
	"github.com/urfave/cli/v2"
)

var brokers = []string{"localhost:9092"}

func runSequentialTest(c *cli.Context) error {
	topic := xid.New().String()

	if err := createTopic(topic, 1); err != nil {
		return err
	}

	messages := generateMessages(topic, 100)

	// Create a producer and produce messages to the topic
	if err := publishMessages(messages); err != nil {
		return err
	}

	h := &handler{
		expect: messages,
	}

	// Create a consumer and consume messages from the topic
	consumer, err := xkafka.NewConsumer(
		"test-seq-consumer",
		h,
		xkafka.Brokers(brokers),
		xkafka.Topics{topic},
		xkafka.ConfigMap{
			"enable.auto.commit": true,
			"auto.offset.reset":  "earliest",
		},
	)
	if err != nil {
		return err
	}

	h.close = func() {
		consumer.Close()
	}

	if err := consumer.Start(context.Background()); err != nil {
		return err
	}

	return nil
}

type handler struct {
	recv   []*xkafka.Message
	expect []*xkafka.Message
	close  func()
}

func (h *handler) Handle(ctx context.Context, msg *xkafka.Message) error {
	slog.Info("[CONSUMER] received message",
		"key", string(msg.Key), "value", string(msg.Value),
		"offset", msg.Offset, "partition", msg.Partition,
	)

	h.recv = append(h.recv, msg)

	if len(h.recv) == len(h.expect) {
		h.close()
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

		slog.Info("[PRODUCER] published message", "key", string(msg.Key), "value", string(msg.Value))
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
