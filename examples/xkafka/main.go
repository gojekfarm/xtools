package main

import (
	"context"
	"fmt"
	"os"
	"sync"

	"github.com/rs/xid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/gojekfarm/xtools/xkafka"
)

const (
	topic  = "xkafka-example"
	broker = "localhost:9092"
)

func main() {
	setupLogger()

	var wg sync.WaitGroup

	ctx := context.Background()

	producer, err := xkafka.NewProducer(
		"xkafka-producer",
		xkafka.Brokers{broker},
	)
	if err != nil {
		log.Fatal().Msgf("%s", err)
	}

	consumer, err := xkafka.NewConsumer(
		"xkafka-consumer",
		handler(&wg),
		xkafka.Brokers{broker},
		xkafka.Topics{topic},
		xkafka.ConfigMap{
			"enable.auto.commit": true,
			"auto.offset.reset":  "earliest",
		},
	)
	if err != nil {
		log.Fatal().Msgf("%s", err)
	}

	publishMessages(producer, &wg)

	go func() {
		if err := consumer.Start(ctx); err != nil {
			log.Fatal().Msgf("Consumer error: %s", err)
		}
	}()

	wg.Wait()
}

func publishMessages(producer *xkafka.Producer, wg *sync.WaitGroup) {
	messages := generateMessages(10)

	for _, message := range messages {
		wg.Add(1)

		log.Info().Msgf("[PRODUCER] Sending message: %s: %s", message.Key, message.Value)

		if err := producer.Publish(context.Background(), message); err != nil {
			log.Fatal().Msgf("%s", err)
		}
	}
}

func handler(wg *sync.WaitGroup) xkafka.Handler {
	return xkafka.HandlerFunc(func(ctx context.Context, msg *xkafka.Message) error {
		defer wg.Done()

		log.Info().Msgf("[CONSUMER] Received message: %s: %s", msg.Key, msg.Value)

		return nil
	})
}

func generateMessages(count int) []*xkafka.Message {
	messages := make([]*xkafka.Message, count)

	for i := 0; i < count; i++ {
		messages[i] = &xkafka.Message{
			Topic: topic,
			Key:   []byte(xid.New().String()),
			Value: []byte(fmt.Sprintf("message-%d", i+1)),
		}
	}

	return messages
}

func setupLogger() {
	zerolog.SetGlobalLevel(zerolog.DebugLevel)

	console := zerolog.ConsoleWriter{
		Out: os.Stderr,
	}

	log.Logger = log.Output(console)
}
