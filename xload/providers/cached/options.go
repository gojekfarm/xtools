package cached

import "time"

// Option configures a cached loader.
type Option interface {
	apply(*options)
}

type optionFunc func(*options)

func (f optionFunc) apply(opts *options) { f(opts) }

// TTL sets the TTL for the cached keys.
type TTL time.Duration

func (t TTL) apply(o *options) { o.ttl = time.Duration(t) }

// Cache sets the cache implementation for the loader.
func Cache(c Cacher) Option {
	return optionFunc(func(o *options) { o.cache = c })
}

// DisableEmptyValueHit disables caching of empty values.
var DisableEmptyValueHit Option = emptyValHit(false)

type emptyValHit bool

func (e emptyValHit) apply(o *options) { o.emptyHit = bool(e) }

type options struct {
	ttl      time.Duration
	cache    Cacher
	emptyHit bool
}

func defaultOptions() *options {
	return &options{
		ttl:      5 * time.Minute,
		emptyHit: true,
	}
}

func (o *options) apply(opts ...Option) {
	for _, opt := range opts {
		opt.apply(o)
	}

	// DESIGN: If no cache is provided, use a simple map cache.
	// This is after applying the options to avoid unnecessary
	// allocations for the default cache.
	if o.cache == nil {
		o.cache = NewMapCache()
	}
}
