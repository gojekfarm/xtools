module github.com/gojekfarm/xtools/examples/xkafka

go 1.25

replace (
	github.com/gojekfarm/xtools => ../../
	github.com/gojekfarm/xtools/xkafka => ../../xkafka
	github.com/gojekfarm/xtools/xkafka/middleware/zerolog => ../../xkafka/middleware/zerolog
)

require (
	github.com/confluentinc/confluent-kafka-go/v2 v2.0.2
	github.com/gojekfarm/xrun v0.3.0
	github.com/gojekfarm/xtools/xkafka v0.11.0
	github.com/gojekfarm/xtools/xkafka/middleware/zerolog v0.10.1
	github.com/rs/xid v1.5.0
	github.com/rs/zerolog v1.29.0
	github.com/urfave/cli/v2 v2.23.7
	golang.org/x/exp v0.0.0-20190121172915-509febef88a4
)

require (
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/hashicorp/errwrap v1.0.0 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.16 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673 // indirect
	go.uber.org/atomic v1.10.0 // indirect
	go.uber.org/multierr v1.9.0 // indirect
	golang.org/x/sys v0.0.0-20220829200755-d48e67d00261 // indirect
)
