package xconfig

import "context"

// Load loads the configuration from environment variables.
func Load(ctx context.Context, cfg any) error {
	return nil
}

// LoadWith loads the configuration using the given options.
func LoadWith(ctx context.Context, cfg any, opts ...option) error {
	o := newOptions(opts...)

	_ = o

	return nil
}
