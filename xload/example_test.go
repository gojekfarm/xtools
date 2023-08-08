package xload_test

import (
	"context"
	"strings"
	"time"

	"github.com/gojekfarm/xtools/xload"
)

func ExampleLoad_default() {
	type AppConf struct {
		Host    string        `env:"HOST"`
		Debug   bool          `env:"DEBUG"`
		Timeout time.Duration `env:"TIMEOUT"`
	}

	var conf AppConf

	err := xload.Load(context.Background(), &conf)
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
		Host    string        `env:"HOST"`
		Debug   bool          `env:"DEBUG"`
		Timeout time.Duration `env:"TIMEOUT"`
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

func ExampleLoad_required() {
	type AppConf struct {
		Host    string        `env:"HOST,required"`
		Debug   bool          `env:"DEBUG"`
		Timeout time.Duration `env:"TIMEOUT"`
	}

	var conf AppConf

	// if HOST is not set, Load will return ErrRequired
	err := xload.Load(context.Background(), &conf)
	if err != nil {
		panic(err)
	}
}

func ExampleLoad_structs() {
	type DBConf struct {
		Host string `env:"HOST"` // will be loaded from DB_HOST
		Port int    `env:"PORT"` // will be loaded from DB_PORT
	}

	type HTTPConf struct {
		Host string `env:"HTTP_HOST"` // will be loaded from HTTP_HOST
		Port int    `env:"HTTP_PORT"` // will be loaded from HTTP_PORT
	}

	type AppConf struct {
		DB   DBConf   `env:",prefix=DB_"` // example of prefix for nested struct
		HTTP HTTPConf // example of embedded struct
	}

	var conf AppConf

	err := xload.Load(context.Background(), &conf)
	if err != nil {
		panic(err)
	}
}

type Host string

func (h *Host) Decode(val string) error {
	// custom decode logic or validation
	return nil
}

func ExampleLoad_customDecoder() {
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
		Host    Host          `env:"HOST"`
		Debug   bool          `env:"DEBUG"`
		Timeout time.Duration `env:"TIMEOUT"`
	}

	var conf AppConf

	err := xload.Load(context.Background(), &conf)
	if err != nil {
		panic(err)
	}
}

func ExampleLoad_transformFieldName() {
	type AppConf struct {
		Host    string        `env:"MYAPP_HOST"`
		Debug   bool          `env:"MYAPP_DEBUG"`
		Timeout time.Duration `env:"MYAPP_TIMEOUT"`
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

func ExampleLoad_arrayDelimiter() {
	type AppConf struct {
		// value will be split by |, instead of ,
		// e.g. HOSTS=host1|host2|host3
		Hosts []string `env:"HOSTS,delimiter=|"`
	}

	var conf AppConf

	err := xload.Load(context.Background(), &conf)
	if err != nil {
		panic(err)
	}
}

func ExampleLoad_mapSeparator() {
	type AppConf struct {
		// key value pair will be split by :, instead of =
		// e.g. HOSTS=db:localhost,cache:localhost
		Hosts map[string]string `env:"HOSTS,separator=:"`
	}

	var conf AppConf

	err := xload.Load(context.Background(), &conf)
	if err != nil {
		panic(err)
	}
}
