package xconfig

// Tag allows customising the struct tag name.
// Default is "config".
type Tag string

func (t Tag) apply(o *options) { o.tag = string(t) }

// CustomLoader allows customising the loader.
func CustomLoader(loader Loader) Option {
	return OptionFunc(func(o *options) { o.loader = loader })
}

type options struct {
	tag    string
	loader Loader
}

func newOptions(opts ...Option) *options {
	o := &options{
		tag:    "config",
		loader: OSLoader(),
	}

	for _, opt := range opts {
		opt.apply(o)
	}

	return o
}

func (o *options) clone() *options {
	return &options{
		tag:    o.tag,
		loader: o.loader,
	}
}

// Option allows customising the behaviour of LoadWith.
type Option interface{ apply(*options) }

// OptionFunc allows functions to be used as options.
type OptionFunc func(*options)

func (f OptionFunc) apply(o *options) { f(o) }

type tagOptions struct {
	prefix    string
	required  bool
	delimiter string
	separator string
}
