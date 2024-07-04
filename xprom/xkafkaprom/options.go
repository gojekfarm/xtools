package xkafkaprom

// Option configures the collector.
type Option interface {
	apply(*options)
}

type optionFunc func(*options)

func (f optionFunc) apply(o *options) { f(o) }

type errorClassifier func(error) string

type options struct {
	latencyBuckets []float64
	errFn          errorClassifier
	address        string
	port           int
}

// LatencyBuckets configures the latency buckets.
type LatencyBuckets []float64

func (l LatencyBuckets) apply(o *options) {
	o.latencyBuckets = l
}

// Address sets `server_address` label.
type Address string

func (a Address) apply(o *options) { o.address = string(a) }

// Port sets `server_port` label.
type Port int

func (p Port) apply(o *options) { o.port = int(p) }

// ErrorClassifer classifies errors types for `error_type` label.
func ErrorClassifer(fn func(error) string) Option {
	return optionFunc(func(o *options) {
		o.errFn = fn
	})
}
