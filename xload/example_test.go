package xload_test

import (
	"context"
	"strings"
	"time"

	"github.com/gojekfarm/xtools/xload"
)

func ExampleLoadEnv() {
	type AppConf struct {
		Host    string        `config:"HOST"`
		Debug   bool          `config:"DEBUG"`
		Timeout time.Duration `config:"TIMEOUT"`
	}

	var conf AppConf

	err := xload.LoadEnv(context.Background(), &conf)
	if err != nil {
		panic(err)
	}
}

func ExampleLoad_customTagNames() {
	type AppConf struct {
		Host    string        `custom:"HOST"`
		Debug   bool          `custom:"DEBUG"`
		Timeout time.Duration `custom:"TIMEOUT"`
	}

	var conf AppConf

	err := xload.Load(context.Background(), &conf, xload.FieldTagName("custom"))
	if err != nil {
		panic(err)
	}
}

func ExampleLoad_customLoader() {
	type AppConf struct {
		Host    string        `env:"HOST"`
		Debug   bool          `env:"DEBUG"`
		Timeout time.Duration `env:"TIMEOUT"`
	}

	var conf AppConf

	loader := xload.LoaderFunc(func(ctx context.Context, key string) (string, error) {
		// lookup value from somewhere
		return "", nil
	})

	err := xload.Load(
		context.Background(),
		&conf,
		xload.FieldTagName("env"),
		xload.WithLoader(loader),
	)
	if err != nil {
		panic(err)
	}
}

func ExampleLoad_prefixLoader() {
	type AppConf struct {
		Host    string        `config:"HOST"`
		Debug   bool          `config:"DEBUG"`
		Timeout time.Duration `config:"TIMEOUT"`
	}

	var conf AppConf

	err := xload.Load(
		context.Background(),
		&conf,
		xload.WithLoader(xload.PrefixLoader("MYAPP_", xload.OSLoader())),
	)
	if err != nil {
		panic(err)
	}
}

func ExampleLoadEnv_required() {
	type AppConf struct {
		Host    string        `config:"HOST,required"`
		Debug   bool          `config:"DEBUG"`
		Timeout time.Duration `config:"TIMEOUT"`
	}

	var conf AppConf

	// if HOST is not set, Load will return ErrRequired
	err := xload.LoadEnv(context.Background(), &conf)
	if err != nil {
		panic(err)
	}
}

func ExampleLoadEnv_structs() {
	type DBConf struct {
		Host string `config:"HOST"` // will be loaded from DB_HOST
		Port int    `config:"PORT"` // will be loaded from DB_PORT
	}

	type HTTPConf struct {
		Host string `config:"HTTP_HOST"` // will be loaded from HTTP_HOST
		Port int    `config:"HTTP_PORT"` // will be loaded from HTTP_PORT
	}

	type AppConf struct {
		DB   DBConf   `config:",prefix=DB_"` // example of prefix for nested struct
		HTTP HTTPConf // example of embedded struct
	}

	var conf AppConf

	err := xload.LoadEnv(context.Background(), &conf)
	if err != nil {
		panic(err)
	}
}

type Host string

func (h *Host) Decode(val string) error {
	// custom decode logic or validation
	return nil
}

func ExampleLoadEnv_customDecoder() {
	// type Host string
	//
	// func (h *Host) Decode(val string) error {
	// 	// custom decode logic or validation
	// 	return nil
	// }
	//
	// Custom decoder can be used for any type that
	// implements the Decoder interface.

	type AppConf struct {
		Host    Host          `config:"HOST"`
		Debug   bool          `config:"DEBUG"`
		Timeout time.Duration `config:"TIMEOUT"`
	}

	var conf AppConf

	err := xload.LoadEnv(context.Background(), &conf)
	if err != nil {
		panic(err)
	}
}

func ExampleLoadEnv_transformFieldName() {
	type AppConf struct {
		Host    string        `config:"MYAPP_HOST"`
		Debug   bool          `config:"MYAPP_DEBUG"`
		Timeout time.Duration `config:"MYAPP_TIMEOUT"`
	}

	var conf AppConf

	// transform converts key from MYAPP_HOST to myapp.host
	transform := func(next xload.Loader) xload.Loader {
		return xload.LoaderFunc(func(ctx context.Context, key string) (string, error) {
			newKey := strings.ReplaceAll(key, "_", ".")
			newKey = strings.ToLower(newKey)

			return next.Load(ctx, newKey)
		})
	}

	err := xload.Load(
		context.Background(),
		&conf,
		xload.WithLoader(transform(xload.OSLoader())),
	)
	if err != nil {
		panic(err)
	}
}
