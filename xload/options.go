package xload

const defaultKey = "env"

// Option configures the xload behaviour.
type Option interface{ apply(*options) }

// FieldTagName allows customising the struct tag name to use.
type FieldTagName string

func (k FieldTagName) apply(opts *options) { opts.tagName = string(k) }

// Concurrency allows customising the number of goroutines to use.
// Default is 1.
type Concurrency int

func (c Concurrency) apply(opts *options) { opts.concurrency = int(c) }

// WithLoader allows customising the loader to use.
func WithLoader(loader Loader) Option {
	return optionFunc(func(opts *options) { opts.loader = loader })
}

// SkipCollisionDetection disables detecting any key collisions while trying to load full keys.
var SkipCollisionDetection = &applier{f: func(o *options) { o.detectCollisions = false }}

// optionFunc allows using a function as an Option.
type optionFunc func(*options)

func (f optionFunc) apply(opts *options) { f(opts) }

type applier struct{ f func(*options) }

func (a *applier) apply(opts *options) { a.f(opts) }

// options holds the configuration.
type options struct {
	tagName          string
	loader           Loader
	concurrency      int
	detectCollisions bool
}

func newOptions(opts ...Option) *options {
	o := &options{
		tagName:          defaultKey,
		loader:           OSLoader(),
		concurrency:      1,
		detectCollisions: true,
	}

	for _, opt := range opts {
		opt.apply(o)
	}

	return o
}
