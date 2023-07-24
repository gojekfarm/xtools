package xconfig

import "context"

// LoadEnv reads config from OS environment using default options.
func LoadEnv(ctx context.Context, v any) error {
	return Load(ctx, v)
}

// Load reads config into the given struct using the given options.
func Load(ctx context.Context, v any, opts ...Option) error {
	o := newOptions(opts...)

	nodes, err := parse(v, o)
	if err != nil {
		return err
	}

	for _, n := range nodes {
		val, err := o.loader.Load(ctx, n.name)
		if err != nil {
			return err
		}

		err = n.SetVal(val)
		if err != nil {
			return err
		}
	}

	return nil
}
