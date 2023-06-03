module github.com/gojekfarm/xtools/examples/xkafka

go 1.19

replace (
	github.com/gojekfarm/xtools => ../../
	github.com/gojekfarm/xtools/xkafka => ../../xkafka
	github.com/gojekfarm/xtools/xkafka/middleware/prometheus => ../../xkafka/middleware/prometheus
	github.com/gojekfarm/xtools/xkafka/middleware/zerolog => ../../xkafka/middleware/zerolog

)

require (
	github.com/gojekfarm/xrun v0.3.0
	github.com/gojekfarm/xtools/xkafka v0.2.0
	github.com/gojekfarm/xtools/xkafka/middleware/prometheus v0.0.0-00010101000000-000000000000
	github.com/gojekfarm/xtools/xkafka/middleware/zerolog v0.0.0-00010101000000-000000000000
	github.com/prometheus/client_golang v1.14.0
	github.com/rs/xid v1.4.0
	github.com/rs/zerolog v1.29.0
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/cockroachdb/errors v1.9.0 // indirect
	github.com/cockroachdb/logtags v0.0.0-20211118104740-dabe8e521a4f // indirect
	github.com/cockroachdb/redact v1.1.3 // indirect
	github.com/confluentinc/confluent-kafka-go v1.9.2 // indirect
	github.com/getsentry/sentry-go v0.13.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/kr/pretty v0.3.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.16 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/prometheus/common v0.37.0 // indirect
	github.com/prometheus/procfs v0.8.0 // indirect
	github.com/rogpeppe/go-internal v1.9.0 // indirect
	github.com/sourcegraph/conc v0.2.0 // indirect
	github.com/sourcegraph/sourcegraph/lib v0.0.0-20221216004406-749998a2ac74 // indirect
	golang.org/x/sys v0.0.0-20220829200755-d48e67d00261 // indirect
	google.golang.org/protobuf v1.28.1 // indirect
)
