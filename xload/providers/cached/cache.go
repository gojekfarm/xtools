package cached

import (
	"sync"
	"time"
)

// Cacher is the interface for custom cache implementations.
//
//go:generate mockery --name Cacher --structname MockCache --filename mock_test.go --outpkg cached --output .
type Cacher interface {
	Get(key string) (string, error)
	Set(key, value string, ttl time.Duration) error
}

// MapCache is a simple cache implementation using a map.
type MapCache struct {
	m   map[string]string
	ttl map[string]time.Time

	now func() time.Time
	mu  sync.Mutex
}

// NewMapCache returns a new MapCache.
func NewMapCache() *MapCache {
	return &MapCache{
		m:   make(map[string]string),
		ttl: make(map[string]time.Time),
		now: time.Now,
	}
}

// Get returns the value for the given key, if cached.
func (c *MapCache) Get(key string) (string, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.isExpired(key) {
		c.delete(key)

		return "", nil
	}

	if v, ok := c.m[key]; ok {
		return v, nil
	}

	return "", nil
}

// Set sets the value for the given key.
func (c *MapCache) Set(key, value string, ttl time.Duration) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.m[key] = value
	c.ttl[key] = c.now().Add(ttl)

	return nil
}

func (c *MapCache) isExpired(key string) bool {
	t, ok := c.ttl[key]
	if !ok {
		return false
	}

	return c.now().After(t)
}

func (c *MapCache) delete(key string) {
	delete(c.m, key)
	delete(c.ttl, key)
}
