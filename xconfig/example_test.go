package xconfig

import (
	"context"
	"time"
)

type Config struct {
	Host  string `yaml:"HOST" config:"HOST"`
	Port  int    `yaml:"PORT" config:"PORT"`
	Retry Retry  `yaml:"RETRY" config:"RETRY"`
}

type Retry struct {
	Max     int           `yaml:"MAX" config:"MAX"`
	Timeout time.Duration `yaml:"TIMEOUT" config:"TIMEOUT"`
}

func DefaultConfig() *Config {
	return &Config{
		Host: "localhost",
		Port: 8080,
		Retry: Retry{
			Max:     3,
			Timeout: 5 * time.Second,
		},
	}
}

func Example() {
	ctx := context.Background()
	cfg := DefaultConfig()

	err := Load(ctx, cfg)
	if err != nil {
		panic(err)
	}
}

func ExampleLoadWith_prefix() {
	ctx := context.Background()
	cfg := DefaultConfig()

	err := LoadWith(ctx, cfg, Prefix("APP"))
	if err != nil {
		panic(err)
	}
}

func ExampleLoadWith_customLoader() {
	ctx := context.Background()
	cfg := DefaultConfig()

	loaders := Multi{
		LoaderFunc(func(ctx context.Context, key string) (string, bool) {
			return "localhost", true
		}),
	}

	err := LoadWith(ctx, cfg, loaders)
	if err != nil {
		panic(err)
	}
}
