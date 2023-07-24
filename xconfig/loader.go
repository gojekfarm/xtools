package xconfig

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

// MapLoader loads values from a map.
// Mostly used for testing.
type MapLoader map[string]string

// Load fetches the value from the map.
func (m MapLoader) Load(ctx context.Context, key string) (string, error) {
	value, ok := m[key]
	if !ok {
		return "", nil
	}

	return value, nil
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
