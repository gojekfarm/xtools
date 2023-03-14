package main

import (
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/gojekfarm/xtools/xkafka"
	"github.com/gojekfarm/xtools/xkafka/middleware/prometheus"
	zlog "github.com/gojekfarm/xtools/xkafka/middleware/zerolog"
)

const (
	topic  = "xkafka-example"
	broker = "localhost:9092"
)

func newProducer() *xkafka.Producer {
	producer, err := xkafka.NewProducer(
		"xkafka-producer",
		xkafka.Brokers{broker},
	)
	if err != nil {
		log.Fatal().Msgf("%s", err)
	}

	producer.Use(
		zlog.LoggingMiddleware(zerolog.InfoLevel),
		prometheus.ProducerMiddleware,
	)

	return producer
}

func newConsumer(handler xkafka.Handler) *xkafka.Consumer {
	consumer, err := xkafka.NewConsumer(
		"xkafka-consumer",
		handler,
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

	consumer.Use(
		zlog.LoggingMiddleware(zerolog.InfoLevel),
		prometheus.ConsumerMiddleware,
	)

	return consumer
}
