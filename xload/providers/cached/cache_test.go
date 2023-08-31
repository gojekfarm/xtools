package cached

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMapCache(t *testing.T) {
	cache := NewMapCache()

	err := cache.Set("foo", "bar", 1*time.Minute)
	assert.NoError(t, err)

	v, err := cache.Get("foo")
	assert.NoError(t, err)
	assert.Equal(t, "bar", v)

	v, err = cache.Get("not-found")
	assert.NoError(t, err)
	assert.Equal(t, "", v)
}

func TestMapCacheExpired(t *testing.T) {
	cache := NewMapCache()
	now := time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)

	cache.now = func() time.Time {
		return now
	}

	err := cache.Set("foo", "bar", 1*time.Minute)
	assert.NoError(t, err)

	cache.now = func() time.Time {
		return now.Add(2 * time.Minute)
	}

	v, err := cache.Get("foo")
	assert.NoError(t, err)
	assert.Equal(t, "", v)
}
