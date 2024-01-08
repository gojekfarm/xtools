package xload

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"testing"
	"time"

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
		LoaderFunc(loader),
	)
	assert.NoError(t, err)

	want := []string{"NAME", "AGE", "BIO", "AVATAR_URL", "AVATAR_HEIGHT", "AVATAR_WIDTH"}

	for _, key := range want {
		_, ok := loaded.Load(key)
		assert.True(t, ok)
	}
}

func TestLoad_Async_Error(t *testing.T) {
	t.Run("cancelled context", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()

		dest := struct {
			Name string `env:"NAME"`
		}{}

		err := Load(ctx, &dest, Concurrency(2),
			LoaderFunc(func(ctx context.Context, key string) (string, error) {
				// simulate a slow loader
				select {
				case <-ctx.Done():
					return "", ctx.Err()
				case <-time.After(100 * time.Millisecond):
					return "name", nil
				}
			}),
		)

		assert.NotEqual(t, "name", dest.Name)
		assert.EqualError(t, err, context.Canceled.Error())
	})

	errMap := map[string]error{
		"NAME":       errors.New("error: NAME"),
		"AGE":        errors.New("error: AGE"),
		"BIO":        errors.New("error: BIO"),
		"NESTED_VAL": errors.New("error: NESTED_VAL"),
	}

	loader := func(ctx context.Context, key string) (string, error) {
		return "", errMap[key]
	}

	cfg := struct {
		Name   string `env:"NAME"`
		Age    int    `env:"AGE"`
		Bio    string `env:"BIO"`
		Nested struct {
			Val string `env:"VAL"`
		} `env:",prefix=NESTED_"`
	}{}

	err := Load(context.Background(), &cfg,
		Concurrency(2),
		LoaderFunc(loader),
	)
	assert.Error(t, err)

	for _, want := range errMap {
		assert.True(t, errors.Is(err, want))
	}
}

type CustomString string

func (cs *CustomString) Decode(s string) error {
	*cs = CustomString(s)
	return nil
}

func Test_loadAndSetWithOriginal(t *testing.T) {
	type Args struct {
		Nest *struct {
			Val CustomString `env:"VAL,required"`
		}
	}

	t.Run("successful load and set", func(t *testing.T) {
		meta := &field{key: "testName", required: true}

		obj := &Args{
			Nest: &struct {
				Val CustomString `env:"VAL,required"`
			}{},
		}

		original := reflect.ValueOf(&obj.Nest.Val)
		fVal := reflect.ValueOf(&obj.Nest.Val).Elem()

		loader := LoaderFunc(func(ctx context.Context, key string) (string, error) {
			return "loadedValue", nil
		})

		err := loadAndSetWithOriginal(loader, meta)(context.Background(), original, fVal, true)
		assert.Nil(t, err)
		assert.EqualValues(t, "loadedValue", string(obj.Nest.Val))
	})

	t.Run("loader returns error", func(t *testing.T) {
		meta := &field{key: "testName", required: true}
		original := reflect.ValueOf(new(string))
		fVal := reflect.ValueOf(new(string))

		err := loadAndSetWithOriginal(LoaderFunc(func(ctx context.Context, key string) (string, error) {
			return "", errors.New("load error")
		}), meta)(context.Background(), original, fVal, true)
		assert.NotNil(t, err)
		assert.Equal(t, "load error", err.Error())
	})

	t.Run("key is required but loader val is empty", func(t *testing.T) {
		meta := &field{key: "testName", required: true}
		original := reflect.ValueOf(new(string))
		fVal := reflect.ValueOf(new(string))

		err := loadAndSetWithOriginal(LoaderFunc(func(ctx context.Context, key string) (string, error) {
			return "", nil
		}), meta)(context.Background(), original, fVal, true)
		assert.NotNil(t, err)

		wantErr := &ErrRequired{}
		assert.ErrorAs(t, err, &wantErr)
		assert.Equal(t, "testName", wantErr.key)
	})
}
