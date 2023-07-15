package xconfig

// Option configures the xconfig behavior.
type Option interface {
	apply(*options)
}

// OptionFunc allows using a function as an Option.
type OptionFunc func(*options)

func (f OptionFunc) apply(opts *options) { f(opts) }

// options holds the configuration.
type options struct {
	key    string
	loader Loader
}

// Key allows customizing the struct tag name to use.
type Key string

func (k Key) apply(opts *options) { opts.key = string(k) }

// WithLoader allows customizing the loader to use.
func WithLoader(loader Loader) Option {
	return OptionFunc(func(opts *options) {
		opts.loader = loader
	})
}
