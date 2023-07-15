package xconfig

import "context"

// LoadEnv reads config from OS environment using default options.
func LoadEnv(ctx context.Context, v any) error {
	return Load(ctx, v)
}

// Load reads config into the given struct using the given options.
func Load(ctx context.Context, v any, opts ...Option) error {
	return nil
}
