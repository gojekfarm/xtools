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

// OSLoader loads values from environment variables.
func OSLoader() Loader {
	return LoaderFunc(func(ctx context.Context, key string) (string, bool) {
		return os.LookupEnv(key)
	})
}

// Multi sequentially tries to load values from a list of loaders.
// First available value is returned.
type Multi []Loader

// Load sequentially tries to load values from a list of loaders.
func (m Multi) Load(ctx context.Context, key string) (string, bool) {
	for _, loader := range m {
		if value, ok := loader.Load(ctx, key); ok {
			return value, true
		}
	}

	return "", false
}

// PrefixLoader wraps a loader and adds a prefix to all keys.
func PrefixLoader(prefix string, loader Loader) Loader {
	return LoaderFunc(func(ctx context.Context, key string) (string, bool) {
		return loader.Load(ctx, prefix+key)
	})
}

// MapLoader loads values from a map.
// Mostly used for testing.
type MapLoader map[string]string

// Load fetches the value from the map.
func (m MapLoader) Load(ctx context.Context, key string) (string, bool) {
	value, ok := m[key]

	return value, ok
}
