package yaml_test

import (
	"context"
	"os"

	"github.com/gojekfarm/xtools/xload"
	"github.com/gojekfarm/xtools/xload/providers/yaml"
)

func Example() {
	// This example shows how to load values from a YAML file.
	//
	// The example YAML file looks like this:
	//
	// NAME: xload
	// VERSION: 1.1

	ctx := context.Background()
	cfg := struct {
		Name    string `env:"NAME"`
		Version string `env:"VERSION"`
	}{}

	f, err := os.Open("example.yaml")
	if err != nil {
		panic(err)
	}

	loader, err := yaml.NewLoader(f, "_")
	if err != nil {
		panic(err)
	}

	err = xload.Load(ctx, &cfg, xload.WithLoader(loader))
	if err != nil {
		panic(err)
	}
}

func Example_nestedStruct() {
	// This example shows how to load values from a YAML file into a nested struct.
	//
	// The example YAML file looks like this:
	//
	// NAME: xload
	// VERSION: 1.1
	// DATABASE:
	//   HOST: localhost
	//   PORT: 5432

	ctx := context.Background()
	cfg := struct {
		Name    string `env:"NAME"`
		Version string `env:"VERSION"`
		DB      struct {
			Host string `env:"HOST"`
			Port int    `env:"PORT"`
		} `env:",prefix=DATABASE_"`
	}{}

	f, err := os.Open("example.yaml")
	if err != nil {
		panic(err)
	}

	// IMPORTANT: The prefix for nested struct is specified in
	// the struct tag. In this case, the prefix is DATABASE_.
	// The separator specified in the NewLoader should match
	// the prefix. In this case, the separator is "_".
	loader, err := yaml.NewLoader(f, "_")
	if err != nil {
		panic(err)
	}

	err = xload.Load(ctx, &cfg, xload.WithLoader(loader))
	if err != nil {
		panic(err)
	}
}
