module github.com/gojekfarm/xtools/e2e/xkafka

go 1.21

replace github.com/gojekfarm/xtools/xkafka => ../../xkafka

require (
	github.com/confluentinc/confluent-kafka-go v1.9.2
	github.com/gojekfarm/xtools/xkafka v0.0.0-00010101000000-000000000000
	github.com/lmittmann/tint v1.0.2
	github.com/rs/xid v1.5.0
	github.com/urfave/cli/v2 v2.25.7
)

require (
	github.com/cockroachdb/errors v1.9.0 // indirect
	github.com/cockroachdb/logtags v0.0.0-20211118104740-dabe8e521a4f // indirect
	github.com/cockroachdb/redact v1.1.3 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/getsentry/sentry-go v0.13.0 // indirect
	github.com/gogo/protobuf v1.3.2 // indirect
	github.com/kr/pretty v0.3.0 // indirect
	github.com/kr/text v0.2.0 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/rogpeppe/go-internal v1.9.0 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sourcegraph/conc v0.2.0 // indirect
	github.com/sourcegraph/sourcegraph/lib v0.0.0-20221216004406-749998a2ac74 // indirect
	github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673 // indirect
	golang.org/x/sys v0.13.0 // indirect
)