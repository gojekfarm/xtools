package xconfig

import (
	"context"
	"os"
)

// Loader defines the interface for custom loaders.
type Loader interface {
	Load(ctx context.Context, key string) (string, bool)
}

// LoaderFunc allows the use of ordinary functions as loaders.
type LoaderFunc func(ctx context.Context, key string) (string, bool)

// Load implements the Loader interface.
func (f LoaderFunc) Load(ctx context.Context, key string) (string, bool) {
	return f(ctx, key)
}

type osLoader struct{}

// Load fetches the value from the environment.
func (o *osLoader) Load(ctx context.Context, key string) (string, bool) {
	return os.LookupEnv(key)
}

func (o *osLoader) apply(opts *options) { opts.loader = o }

// OsLoader returns a loader that loads values from environment variables.
func OsLoader() *osLoader {
	return &osLoader{}
}

type Multi []Loader

func (m Multi) Load(ctx context.Context, key string) (string, bool) {
	for _, loader := range m {
		if value, ok := loader.Load(ctx, key); ok {
			return value, true
		}
	}

	return "", false
}

func (m Multi) apply(o *options) { o.loader = m }
