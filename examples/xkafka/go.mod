module github.com/gojekfarm/xtools/examples/xkafka

go 1.21

replace (
	github.com/gojekfarm/xtools => ../../
	github.com/gojekfarm/xtools/xkafka => ../../xkafka

)

require (
	github.com/confluentinc/confluent-kafka-go v1.9.2
	github.com/gojekfarm/xrun v0.3.0
	github.com/gojekfarm/xtools/xkafka v0.6.0
	github.com/lmittmann/tint v1.0.3
	github.com/rs/xid v1.4.0
	github.com/urfave/cli/v2 v2.23.7
)

require (
	github.com/confluentinc/confluent-kafka-go/v2 v2.0.2 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673 // indirect
	go.uber.org/atomic v1.10.0 // indirect
	go.uber.org/multierr v1.9.0 // indirect
)
