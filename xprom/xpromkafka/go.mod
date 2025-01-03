module github.com/gojekfarm/xtools/xprom/xpromkafka

go 1.21

replace (
	github.com/gojekfarm/xtools/xkafka => ../../xkafka
	github.com/gojekfarm/xtools/xprom/semconv => ../semconv
)

require (
	github.com/gojekfarm/xtools/xkafka v0.9.0
	github.com/gojekfarm/xtools/xprom/semconv v0.9.0
	github.com/prometheus/client_golang v1.14.0
	github.com/stretchr/testify v1.8.1
)

require (
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/confluentinc/confluent-kafka-go/v2 v2.0.2 // indirect
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/matttproud/golang_protobuf_extensions v1.0.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	github.com/prometheus/client_model v0.3.0 // indirect
	github.com/prometheus/common v0.37.0 // indirect
	github.com/prometheus/procfs v0.8.0 // indirect
	github.com/rs/xid v1.5.0 // indirect
	github.com/sourcegraph/conc v0.3.0 // indirect
	go.uber.org/atomic v1.10.0 // indirect
	go.uber.org/multierr v1.9.0 // indirect
	golang.org/x/sys v0.0.0-20220829200755-d48e67d00261 // indirect
	google.golang.org/protobuf v1.33.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)
