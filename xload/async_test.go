package xload

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoad_Async(t *testing.T) {
	loaded := sync.Map{}

	loader := func(ctx context.Context, key string) (string, error) {
		loaded.Store(key, true)

		return "", nil
	}

	cfg := struct {
		Name   string `env:"NAME"`
		Age    int    `env:"AGE"`
		Bio    string `env:"BIO"`
		Avatar struct {
			URL    string `env:"URL"`
			Height int    `env:"HEIGHT"`
			Width  int    `env:"WIDTH"`
		} `env:",prefix=AVATAR_"`
	}{}

	err := Load(context.Background(), &cfg,
		Concurrency(3),
		WithLoader(LoaderFunc(loader)),
	)
	assert.NoError(t, err)

	want := []string{"NAME", "AGE", "BIO", "AVATAR_URL", "AVATAR_HEIGHT", "AVATAR_WIDTH"}

	for _, key := range want {
		_, ok := loaded.Load(key)
		assert.True(t, ok)
	}
}

func TestLoad_Async_Error(t *testing.T) {
	errMap := map[string]error{
		"NAME": errors.New("error: NAME"),
		"AGE":  errors.New("error: AGE"),
		"BIO":  errors.New("error: BIO"),
	}

	loader := func(ctx context.Context, key string) (string, error) {
		return "", errMap[key]
	}

	cfg := struct {
		Name string `env:"NAME"`
		Age  int    `env:"AGE"`
		Bio  string `env:"BIO"`
	}{}

	err := Load(context.Background(), &cfg,
		Concurrency(2),
		WithLoader(LoaderFunc(loader)),
	)
	assert.Error(t, err)

	for _, want := range errMap {
		assert.True(t, errors.Is(err, want))
	}
}
