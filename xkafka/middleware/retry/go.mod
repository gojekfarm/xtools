module github.com/gojekfarm/xtools/xkafka/middleware/retry

go 1.22

replace github.com/gojekfarm/xtools/xkafka => ../../

require (
	github.com/cenkalti/backoff/v4 v4.3.0
	github.com/gojekfarm/xtools/xkafka v0.0.0-00010101000000-000000000000
)

require (
	github.com/confluentinc/confluent-kafka-go/v2 v2.0.2 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	go.uber.org/atomic v1.10.0 // indirect
	go.uber.org/multierr v1.9.0 // indirect
)
