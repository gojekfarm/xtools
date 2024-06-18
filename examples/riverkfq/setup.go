package main

import (
	"context"

	"log/slog"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

func createTopic(name string, partitions int) error {
	admin, err := kafka.NewAdminClient(&kafka.ConfigMap{
		"bootstrap.servers": "localhost:9092",
	})
	if err != nil {
		return err
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
		return err
	}

	slog.Info("[ADMIN] created topic", "name", name, "partitions", partitions, "result", res)

	return nil
}
