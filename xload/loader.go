package xload

import (
	"context"
	"os"
)

// Loader defines the interface for a config loader.
type Loader interface {
	Load(ctx context.Context, key string) (string, error)
}

// LoaderFunc is a function that implements Loader.
type LoaderFunc func(ctx context.Context, key string) (string, error)

// Load returns the config value for the given key.
func (f LoaderFunc) Load(ctx context.Context, key string) (string, error) {
	return f(ctx, key)
}

// PrefixLoader wraps a loader and adds a prefix to all keys.
func PrefixLoader(prefix string, loader Loader) Loader {
	return LoaderFunc(func(ctx context.Context, key string) (string, error) {
		return loader.Load(ctx, prefix+key)
	})
}

// OSLoader loads values from the OS environment.
func OSLoader() Loader {
	return LoaderFunc(func(ctx context.Context, key string) (string, error) {
		v, ok := os.LookupEnv(key)
		if !ok {
			return "", nil
		}

		return v, nil
	})
}

// SerialLoader loads values from multiple loaders.
// Last non-empty value wins.
func SerialLoader(loaders ...Loader) Loader {
	return LoaderFunc(func(ctx context.Context, key string) (string, error) {
		var lastNonEmpty string

		for _, loader := range loaders {
			v, err := loader.Load(ctx, key)
			if err != nil {
				return "", err
			}

			if v != "" {
				lastNonEmpty = v
			}
		}

		return lastNonEmpty, nil
	})
}
