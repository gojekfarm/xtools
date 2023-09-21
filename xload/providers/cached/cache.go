package cached

import (
	"sync"
	"time"
)

// Cacher is the interface for custom cache implementations.
// Implementations must distinguish between key not found and
// key found with empty value.
// `Get` must return nil for key not found.
// `Set` must cache empty values.
//
//go:generate mockery --name Cacher --structname MockCache --filename mock_test.go --outpkg cached --output .
type Cacher interface {
	Get(key string) (*string, error)
	Set(key, value string, ttl time.Duration) error
}

type mv struct {
	val string
	ttl time.Time
}

// MapCache is a simple cache implementation using a map.
type MapCache struct {
	m sync.Map

	now func() time.Time
}

// NewMapCache returns a new MapCache.
func NewMapCache() *MapCache {
	return &MapCache{
		m:   sync.Map{},
		now: time.Now,
	}
}

// Get returns the value for the given key, if cached.
// If the value is not cached, it returns nil.
func (c *MapCache) Get(key string) (*string, error) {
	v, ok := c.m.Load(key)
	if !ok {
		return nil, nil
	}

	mv, ok := v.(*mv)

	if !ok || c.now().After(mv.ttl) {
		c.delete(key)

		return nil, nil
	}

	return &mv.val, nil
}

// Set sets the value for the given key.
func (c *MapCache) Set(key, value string, ttl time.Duration) error {
	v := &mv{
		val: value,
		ttl: c.now().Add(ttl),
	}

	c.m.Store(key, v)

	return nil
}

func (c *MapCache) delete(key string) {
	c.m.Delete(key)
}
