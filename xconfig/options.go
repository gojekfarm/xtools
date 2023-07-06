package xconfig

// Prefix sets the prefix for all configuration keys.
type Prefix string

func (p Prefix) apply(o *options) { o.prefix = string(p) }

type options struct {
	prefix string
	loader Loader
}

func newOptions(opts ...option) *options {
	o := &options{
		loader: OsLoader(),
	}

	for _, opt := range opts {
		opt.apply(o)
	}

	return o
}

type option interface{ apply(*options) }
