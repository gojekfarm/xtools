package xload

import (
	"context"
	"time"
)

func ExampleLoadEnv() {
	type AppConf struct {
		Host    string        `config:"HOST"`
		Debug   bool          `config:"DEBUG"`
		Timeout time.Duration `config:"TIMEOUT"`
	}

	var conf AppConf

	err := LoadEnv(context.Background(), &conf)
	if err != nil {
		panic(err)
	}
}

func ExampleLoad_customTagNames() {
	type AppConf struct {
		Host    string        `env:"HOST"`
		Debug   bool          `env:"DEBUG"`
		Timeout time.Duration `env:"TIMEOUT"`
	}

	var conf AppConf

	err := Load(context.Background(), &conf, FieldTagName("env"))
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

	loader := LoaderFunc(func(ctx context.Context, key string) (string, error) {
		// lookup value from somewhere
		return "", nil
	})

	err := Load(
		context.Background(),
		&conf,
		FieldTagName("env"),
		WithLoader(loader),
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

	loader := LoaderFunc(func(ctx context.Context, key string) (string, error) {
		// lookup value from somewhere
		return "", nil
	})

	err := Load(
		context.Background(),
		&conf,
		WithLoader(PrefixLoader("MYAPP_", loader)),
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
	err := LoadEnv(context.Background(), &conf)
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

	err := LoadEnv(context.Background(), &conf)
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

	err := LoadEnv(context.Background(), &conf)
	if err != nil {
		panic(err)
	}
}
