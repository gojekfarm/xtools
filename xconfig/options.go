package xconfig

import "context"

// Loader defines the interface for custom loaders.
type Loader interface {
	Load(ctx context.Context, key string) ([]byte, error)
}

// LoaderFunc allows the use of ordinary functions as loaders.
type LoaderFunc func(ctx context.Context, key string) ([]byte, error)

// Load implements the Loader interface.
func (f LoaderFunc) Load(ctx context.Context, key string) ([]byte, error) {
	return f(ctx, key)
}

// Prefix sets the prefix for all configuration keys.
type Prefix string

func (p Prefix) apply(o *options) { o.prefix = string(p) }

// Loaders sets the loaders for the configuration.
type Loaders []Loader

func (l Loaders) apply(o *options) {
	o.loaders = append(o.loaders, l...)
}

type options struct {
	prefix  string
	loaders []Loader
}

func newOptions(opts ...option) *options {
	o := new(options)

	for _, opt := range opts {
		opt.apply(o)
	}

	return o
}

type option interface{ apply(*options) }
