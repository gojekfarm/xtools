package cached

import (
	"context"

	"github.com/gojekfarm/xtools/xload"
)

// NewLoader returns a new cached loader.
func NewLoader(l xload.Loader, opts ...Option) xload.LoaderFunc {
	o := defaultOptions()

	o.apply(opts...)

	return xload.LoaderFunc(func(ctx context.Context, key string) (string, error) {
		v, err := o.cache.Get(key)
		if err != nil {
			return "", err
		}

		if v != "" {
			return v, nil
		}

		loaded, err := l.Load(ctx, key)
		if err != nil {
			return "", err
		}

		err = o.cache.Set(key, loaded, o.ttl)

		return loaded, err
	})
}
