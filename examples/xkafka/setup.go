package main

import (
	"context"
	"fmt"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
	"github.com/rs/xid"
	"log/slog"

	"github.com/gojekfarm/xtools/xkafka"
)

var brokers = []string{"localhost:9092"}

func createTopic(partitions int) string {
	name := "xkafka-" + xid.New().String()

	admin, err := kafka.NewAdminClient(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
	})
	if err != nil {
		panic(err)
	}

	res, err := admin.CreateTopics(
		context.Background(),
		[]kafka.TopicSpecification{{
			Topic:             name,
			NumPartitions:     partitions,
			ReplicationFactor: 1,
		}},
	)
	if err != nil {
		panic(err)
	}

	slog.Info("[ADMIN] created topic", "name", name, "partitions", partitions, "result", res)

	return name
}

func generateMessages(topic string, count int) []*xkafka.Message {
	messages := make([]*xkafka.Message, count)

	for i := 0; i < count; i++ {
		messages[i] = &xkafka.Message{
			Topic: topic,
			Key:   []byte(fmt.Sprintf("key-%d", i)),
			Value: []byte(fmt.Sprintf("value-%d : %s", i, xid.New().String())),
		}
	}

	return messages
}
