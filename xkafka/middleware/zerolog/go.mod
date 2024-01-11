module github.com/gojekfarm/xtools/xkafka/middleware/zerolog

go 1.21

toolchain go1.21.0

replace github.com/gojekfarm/xtools/xkafka => ../../

require (
	github.com/gojekfarm/xtools/xkafka v0.6.0
	github.com/rs/zerolog v1.29.0
)

require (
	github.com/confluentinc/confluent-kafka-go/v2 v2.0.2 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.16 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	go.uber.org/atomic v1.10.0 // indirect
	go.uber.org/multierr v1.9.0 // indirect
	golang.org/x/sys v0.0.0-20220829200755-d48e67d00261 // indirect
)
