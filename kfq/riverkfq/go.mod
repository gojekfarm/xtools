module github.com/gojekfarm/xtools/kfq/riverkfq

go 1.21.4

replace github.com/gojekfarm/xtools/xkafka => ../../xkafka

require (
	github.com/gojekfarm/xtools/xkafka v0.9.0
	github.com/jackc/pgx/v5 v5.6.0
	github.com/riverqueue/river v0.7.0
	github.com/riverqueue/river/riverdriver/riverpgxv5 v0.7.0
	github.com/stretchr/testify v1.9.0
)

require (
	github.com/confluentinc/confluent-kafka-go/v2 v2.0.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/riverqueue/river/riverdriver v0.7.0 // indirect
	github.com/riverqueue/river/rivertype v0.7.0 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	go.uber.org/atomic v1.10.0 // indirect
	go.uber.org/multierr v1.9.0 // indirect
	golang.org/x/crypto v0.17.0 // indirect
	golang.org/x/sync v0.7.0 // indirect
	golang.org/x/text v0.16.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
