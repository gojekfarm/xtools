module github.com/gojekfarm/xtools/xkafka/middleware/retry

go 1.21

replace github.com/gojekfarm/xtools/xkafka => ../../

require (
	github.com/cenkalti/backoff/v4 v4.3.0
	github.com/gojekfarm/xtools/xkafka v0.9.0
	github.com/stretchr/testify v1.8.1
)

require (
	github.com/confluentinc/confluent-kafka-go/v2 v2.0.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	go.uber.org/atomic v1.10.0 // indirect
	go.uber.org/multierr v1.9.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
