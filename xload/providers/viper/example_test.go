package viper_test

import (
	"context"
	"fmt"

	vpr "github.com/spf13/viper"

	"github.com/gojekfarm/xtools/xload"
	"github.com/gojekfarm/xtools/xload/providers/viper"
)

type Config struct {
	Log LogConfig `env:",prefix=LOG_"`
}

type LogConfig struct {
	Level string `env:"LEVEL"`
}

func ExampleNew() {
	v := vpr.New()
	v.Set("LOG_LEVEL", "debug")

	vl, err := viper.New(viper.From(v))
	if err != nil {
		panic(err)
	}

	cfg := &Config{}
	if err := xload.Load(context.Background(), cfg, xload.WithLoader(vl)); err != nil {
		panic(err)
	}

	fmt.Println(cfg.Log.Level)

	// Output:
	// debug
}
