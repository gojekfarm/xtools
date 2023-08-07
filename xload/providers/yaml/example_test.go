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
