package xload

const defaultKey = "config"

// option configures the xload behaviour.
type option interface {
	apply(*options)
}

// optionFunc allows using a function as an option.
type optionFunc func(*options)

func (f optionFunc) apply(opts *options) { f(opts) }

// options holds the configuration.
type options struct {
	tagName string
	loader  Loader
}

func newOptions(opts ...option) *options {
	o := &options{
		tagName: defaultKey,
		loader:  OSLoader(),
	}

	for _, opt := range opts {
		opt.apply(o)
	}

	return o
}

// FieldTagName allows customising the struct tag name to use.
type FieldTagName string

func (k FieldTagName) apply(opts *options) { opts.tagName = string(k) }

// WithLoader allows customising the loader to use.
//
//nolint:golint
func WithLoader(loader Loader) option {
	return optionFunc(func(opts *options) {
		opts.loader = loader
	})
}
