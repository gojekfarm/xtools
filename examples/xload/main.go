package main

import (
	"context"
	"encoding/json"
	"time"

	"github.com/gojekfarm/xtools/xload"
	"github.com/gojekfarm/xtools/xload/providers/yaml"
)

func businessConfigLoader() xload.Loader {
	return xload.LoaderFunc(func(ctx context.Context, key string) (string, error) {
		// fetch from remote config
		return "", nil
	})
}

func main() {
	ctx := context.Background()

	// application config is loaded once at startup
	// from yaml, and then from environment variables
	cfg := newAppConfig()

	yamlLoader, _ := yaml.NewFileLoader("application.yaml", "_")

	err := xload.Load(
		ctx,
		cfg,
		xload.FieldTagName("env"),
		xload.WithLoader(
			xload.SerialLoader(
				yamlLoader,
				xload.OSLoader(),
			),
		),
	)
	if err != nil {
		panic(err)
	}

	prettyPrint(cfg)

	// request config is loaded for every request
	// defaults are set in the code, and then
	// overridden by remote config.
	//
	// This can be used in HTTP, GRPC, handlers or
	// middleware.
	rcfg := newRequestConfig()

	reqYamlLoader, err := yaml.NewFileLoader("request.default.yaml", "_")
	if err != nil {
		panic(err)
	}

	err = xload.Load(
		ctx,
		rcfg,
		xload.FieldTagName("env"),
		xload.WithLoader(
			xload.SerialLoader(
				reqYamlLoader,
				businessConfigLoader(),
			),
		),
	)
	if err != nil {
		panic(err)
	}

	prettyPrint(rcfg)
}

func prettyPrint(v interface{}) {
	out, _ := json.MarshalIndent(v, "", "  ")
	println(string(out))
}

func newAppConfig() *AppConfig {
	return &AppConfig{
		Name: "xload",
		HTTP: HTTPConfig{
			Host:  "localhost",
			Port:  8080,
			Debug: true,
		},
		GRPC: &GRPCConfig{
			Host:  "localhost",
			Port:  8081,
			Debug: true,
		},
		DB: DBConfig{
			URL: "localhost:5432",
		},
	}
}

func newRequestConfig() *RequestConfig {
	return &RequestConfig{
		Timeout:     5 * time.Second,
		FeatureFlag: true,
		LaunchV3: &ExperimentConfig{
			Enabled:      false,
			SearchRadius: 2.5,
		},
	}
}
