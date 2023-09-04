package cached

import (
	"context"

	"github.com/gojekfarm/xtools/xload"
)

// NewLoader returns a new cached loader.
//
// The cached loader uses these defaults:
// - TTL: 5 minutes. Configurable via `TTL` option.
// - Cache: A simple unbounded map cache. Configurable via `Cache` option.
// - Empty value hit: Enabled. Configurable via `DisableEmptyValueHit` option.
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

		if loaded == "" && !o.emptyValueHit {
			return "", nil
		}

		err = o.cache.Set(key, loaded, o.ttl)

		return loaded, err
	})
}
