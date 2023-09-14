package cached

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/gojekfarm/xtools/xload"
)

type config struct {
	Key1 string `env:"KEY_1"`
	Key2 string `env:"KEY_2"`
	Key3 string `env:"KEY_3"`
}

func TestNewLoader(t *testing.T) {
	loader := xload.MapLoader(map[string]string{
		"KEY_1": "value-1",
		"KEY_2": "value-2",
	})

	cachedLoader := NewLoader(loader)

	t.Run("Cache MISS", func(t *testing.T) {
		cfg := config{}

		err := xload.Load(context.TODO(), &cfg, cachedLoader)
		assert.NoError(t, err)
		assert.Equal(t, "value-1", cfg.Key1)
		assert.Equal(t, "value-2", cfg.Key2)
	})

	t.Run("Cache HIT", func(t *testing.T) {
		cfg := config{}

		err := xload.Load(context.TODO(), &cfg, cachedLoader)
		assert.NoError(t, err)
		assert.Equal(t, "value-1", cfg.Key1)
		assert.Equal(t, "value-2", cfg.Key2)
	})
}

func TestNewLoader_WithTTL(t *testing.T) {
	loader := xload.MapLoader(map[string]string{
		"KEY_1": "value-1",
		"KEY_2": "value-2",
	})

	ttl := 123 * time.Second

	mc := NewMockCache(t)
	mc.On("Get", mock.Anything).Return(nil, nil).Times(3)
	mc.On("Set", mock.Anything, mock.Anything, ttl).Return(nil).Times(3)

	cachedLoader := NewLoader(loader, TTL(ttl), Cache(mc))

	cfg := config{}

	err := xload.Load(context.Background(), &cfg, cachedLoader)
	assert.NoError(t, err)

	mc.AssertExpectations(t)
}

func TestNewLoader_EmptyValueHit(t *testing.T) {
	loader := xload.MapLoader(map[string]string{
		"KEY_1": "value-1",
		"KEY_2": "value-2",
	})

	ttl := 123 * time.Second

	mc := NewMockCache(t)
	mc.On("Get", "KEY_1").Return(nil, nil)
	mc.On("Set", "KEY_1", "value-1", ttl).Return(nil)

	mc.On("Get", "KEY_2").Return(nil, nil)
	mc.On("Set", "KEY_2", "value-2", ttl).Return(nil)

	mc.On("Get", "KEY_3").Return(nil, nil)
	mc.On("Set", "KEY_3", "", ttl).Return(nil)

	cachedLoader := NewLoader(loader, TTL(ttl), Cache(mc))

	cfg := config{}

	err := xload.Load(context.Background(), &cfg, cachedLoader)
	assert.NoError(t, err)

	// load again to ensure that the empty value is cached
	mc.On("Get", "KEY_1").Return("value-1", nil)
	mc.On("Get", "KEY_2").Return("value-2", nil)
	mc.On("Get", "KEY_3").Return("", nil)

	err = xload.Load(context.Background(), &cfg, cachedLoader)
	assert.NoError(t, err)

	mc.AssertExpectations(t)
}

func TestNewLoader_WithDisableEmptyValueHit(t *testing.T) {
	loader := xload.MapLoader(map[string]string{
		"KEY_1": "value-1",
		"KEY_2": "value-2",
	})

	ttl := 123 * time.Second

	mc := NewMockCache(t)
	mc.On("Get", "KEY_1").Return(nil, nil)
	mc.On("Set", "KEY_1", "value-1", ttl).Return(nil)

	mc.On("Get", "KEY_2").Return(nil, nil)
	mc.On("Set", "KEY_2", "value-2", ttl).Return(nil)

	mc.On("Get", "KEY_3").Return(nil, nil)

	cachedLoader := NewLoader(loader, TTL(ttl), Cache(mc), DisableEmptyValueHit)

	cfg := config{}

	err := xload.Load(context.Background(), &cfg, cachedLoader)
	assert.NoError(t, err)

	// load again to ensure that the empty value is not cached
	mc.On("Get", "KEY_1").Return("value-1", nil)
	mc.On("Get", "KEY_2").Return("value-2", nil)
	mc.On("Get", "KEY_3").Return(nil, nil)

	err = xload.Load(context.Background(), &cfg, cachedLoader)
	assert.NoError(t, err)

	mc.AssertExpectations(t)
}

func TestNewLoader_ForwardError(t *testing.T) {
	failingLoader := xload.LoaderFunc(func(ctx context.Context, key string) (string, error) {
		return "", assert.AnError
	})

	cachedLoader := NewLoader(failingLoader)

	cfg := config{}

	err := xload.Load(context.Background(), &cfg, cachedLoader)
	assert.Error(t, err)
	assert.Equal(t, assert.AnError, err)
}

func TestNewLoader_CacheError(t *testing.T) {
	loader := xload.MapLoader(map[string]string{
		"KEY_1": "value-1",
		"KEY_2": "value-2",
	})

	cfg := config{}

	t.Run("Cache SET error", func(t *testing.T) {
		mc := NewMockCache(t)

		mc.On("Get", "KEY_1").Return(nil, nil)
		mc.On("Set", "KEY_1", "value-1", mock.Anything).Return(assert.AnError)

		cachedLoader := NewLoader(loader, Cache(mc))

		err := xload.Load(context.Background(), &cfg, cachedLoader)
		assert.Error(t, err)
		assert.Equal(t, assert.AnError, err)

		mc.AssertExpectations(t)
	})

	t.Run("Cache GET error", func(t *testing.T) {
		mc := NewMockCache(t)

		mc.On("Get", "KEY_1").Return(nil, assert.AnError)

		cachedLoader := NewLoader(loader, Cache(mc))

		err := xload.Load(context.Background(), &cfg, cachedLoader)
		assert.Error(t, err)
		assert.Equal(t, assert.AnError, err)

		mc.AssertExpectations(t)
	})
}
