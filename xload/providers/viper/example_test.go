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

func ExampleFrom() {
	_, err := viper.New(viper.From(vpr.New()))
	if err != nil {
		panic(err)
	}
}

func ExampleLoader_ConfigFileUsed() {
	_, err := viper.New(viper.ConfigFile("<path-to-config-file>.ext"))
	if err != nil {
		panic(err)
	}

	fmt.Println(vpr.ConfigFileUsed()) // <path-to-config-file>.ext
}

func ExampleNew_fileOptions() {
	_, err := viper.New(
		viper.ConfigName("config"),
		viper.ConfigType("toml"),
		viper.ConfigPaths([]string{"./", "/etc/<program/"}),
	)
	if err != nil {
		panic(err)
	}
}

func ExampleAutoEnv_disable() {
	_, err := viper.New(viper.AutoEnv(false))
	if err != nil {
		panic(err)
	}
}

func ExampleValueMapper() {
	_, err := viper.New(viper.ValueMapper(func(m map[string]any) map[string]string {
		return xload.FlattenMap(m, "__")
	}))
	if err != nil {
		panic(err)
	}
}

func ExampleTransformer() {
	_, err := viper.New(viper.Transformer(func(v *vpr.Viper, next xload.Loader) xload.Loader {
		return xload.PrefixLoader("ENV_", next)
	}))
	if err != nil {
		panic(err)
	}
}
