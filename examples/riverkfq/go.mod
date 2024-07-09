module github.com/gojekfarm/xtools/examples/riverkfq

go 1.21.4

toolchain go1.22.2

replace (
	github.com/gojekfarm/xtools => ../../
	github.com/gojekfarm/xtools/kfq/riverkfq => ../../kfq/riverkfq
	github.com/gojekfarm/xtools/xkafka => ../../xkafka
	github.com/gojekfarm/xtools/xkafka/middleware => ../../xkafka/middleware
)

require (
	github.com/confluentinc/confluent-kafka-go/v2 v2.0.2
	github.com/gojekfarm/xtools/kfq/riverkfq v0.0.0-00010101000000-000000000000
	github.com/gojekfarm/xtools/xkafka v0.8.1
	github.com/gojekfarm/xtools/xkafka/middleware v0.0.0-00010101000000-000000000000
	github.com/jackc/pgx/v5 v5.6.0
	github.com/lmittmann/tint v1.0.4
	github.com/riverqueue/river v0.7.0
	github.com/riverqueue/river/riverdriver/riverpgxv5 v0.7.0
	github.com/rs/xid v1.5.0
)

require (
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20221227161230-091c0ba34f0a // indirect
	github.com/jackc/puddle/v2 v2.2.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/riverqueue/river/riverdriver v0.7.0 // indirect
	github.com/riverqueue/river/rivertype v0.7.0 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	go.uber.org/atomic v1.10.0 // indirect
	go.uber.org/multierr v1.9.0 // indirect
	golang.org/x/crypto v0.17.0 // indirect
	golang.org/x/sync v0.7.0 // indirect
	golang.org/x/text v0.16.0 // indirect
)
