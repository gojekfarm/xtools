package cached_test

import (
	"context"
	"time"

	"github.com/gojekfarm/xtools/xload"
	"github.com/gojekfarm/xtools/xload/providers/cached"
)

func Example() {
	// This example shows how to use the cached loader
	// with a remote loader.

	ctx := context.Background()
	cfg := struct {
		Title       string `env:"TITLE"`
		Link        string `env:"LINK"`
		ButtonLabel string `env:"BUTTON_LABEL"`
	}{}

	remoteLoader := xload.LoaderFunc(func(ctx context.Context, key string) (string, error) {
		// Load the value from a remote source.

		return "", nil
	})

	err := xload.Load(
		ctx, &cfg,
		cached.NewLoader(
			remoteLoader,
			cached.TTL(5*60*time.Minute),
		),
	)
	if err != nil {
		panic(err)
	}
}

type CustomCache struct{}

func NewCustomCache() *CustomCache {
	return &CustomCache{}
}

func (c *CustomCache) Get(key string) (string, error) {
	return "", nil
}

func (c *CustomCache) Set(key, value string, ttl time.Duration) error {
	return nil
}

func Example_customCache() {
	// This example shows how to use a custom cache
	// with the cached loader.

	ctx := context.Background()
	cfg := struct {
		Title       string `env:"TITLE"`
		Link        string `env:"LINK"`
		ButtonLabel string `env:"BUTTON_LABEL"`
	}{}

	remoteLoader := xload.LoaderFunc(func(ctx context.Context, key string) (string, error) {
		// Load the value from a remote source.

		return "", nil
	})

	err := xload.Load(
		ctx, &cfg,
		cached.NewLoader(
			remoteLoader,
			cached.TTL(5*60*time.Minute),
			cached.Cache(NewCustomCache()),
		),
	)
	if err != nil {
		panic(err)
	}
}

func Example_disableEmptyValueHit() {
	// By default, the cached loader caches empty values.
	// This example shows how to disable caching of empty values
	// with the cached loader.

	ctx := context.Background()
	cfg := struct {
		Title       string `env:"TITLE"`
		Link        string `env:"LINK"`
		ButtonLabel string `env:"BUTTON_LABEL"`
	}{}

	remoteLoader := xload.LoaderFunc(func(ctx context.Context, key string) (string, error) {
		// Load the value from a remote source.

		return "", nil
	})

	err := xload.Load(
		ctx, &cfg,
		cached.NewLoader(
			remoteLoader,
			cached.TTL(5*60*time.Minute),
			cached.DisableEmptyValueHit(),
		),
	)
	if err != nil {
		panic(err)
	}
}
