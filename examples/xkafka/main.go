package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"sync"

	goprom "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rs/xid"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"

	"github.com/gojekfarm/xrun"
	"github.com/gojekfarm/xrun/component"
	"github.com/gojekfarm/xtools/xkafka"
	"github.com/gojekfarm/xtools/xkafka/middleware/prometheus"
)

func main() {
	setupLogger()

	promServer := newPrometheus()

	var wg sync.WaitGroup

	ctx, cancel := context.WithCancel(context.Background())

	producer := newProducer()

	consumer := newConsumer(handler(&wg))

	publishMessages(producer, &wg)

	runComponents := func(ctx context.Context) {
		err := xrun.All(xrun.NoTimeout,
			consumer,
			producer,
			component.HTTPServer(
				component.HTTPServerOptions{Server: promServer},
			),
		).Run(ctx)
		if err != nil {
			log.Fatal().Msgf("%s", err)
		}
	}

	go runComponents(ctx)

	// wait for all messages to be consumed
	wg.Wait()

	// stop the consumer and producer
	cancel()
}

func publishMessages(producer *xkafka.Producer, wg *sync.WaitGroup) {
	messages := generateMessages(10)

	for _, message := range messages {
		wg.Add(1)

		log.Info().Msgf("[PRODUCER] Sending message: %s: %s", message.Key, message.Value)

		// add a callback to log the status of the message
		message.AddCallback(func(m *xkafka.Message) {
			log.Info().Msgf("[PRODUCER] Message %s Status %s", m.Key, m.Status)
		})

		if err := producer.Publish(context.Background(), message); err != nil {
			log.Fatal().Msgf("%s", err)
		}
	}
}

func handler(wg *sync.WaitGroup) xkafka.Handler {
	return xkafka.HandlerFunc(func(ctx context.Context, msg *xkafka.Message) error {
		defer wg.Done()

		log.Info().Msgf("[CONSUMER] Received message: %s: %s", msg.Key, msg.Value)

		msg.AckSuccess()

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

func newPrometheus() *http.Server {
	reg := goprom.NewRegistry()

	if err := prometheus.RegisterConsumerMetrics(reg); err != nil {
		log.Fatal().Msgf("%s", err)
	}

	if err := prometheus.RegisterProducerMetrics(reg); err != nil {
		log.Fatal().Msgf("%s", err)
	}

	h := http.NewServeMux()

	h.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{Registry: reg}))

	server := http.Server{
		Addr:    ":9090",
		Handler: h,
	}

	return &server
}
