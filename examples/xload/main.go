package main

import (
	"context"
	"encoding/json"
	"os"
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

	f, err := os.Open("application.yaml")
	if err != nil {
		panic(err)
	}

	yamlLoader, err := yaml.NewLoader(f, "_")
	if err != nil {
		panic(err)
	}

	err = xload.Load(
		ctx,
		cfg,
		xload.FieldTagName("env"),
		xload.WithLoader(
			xload.SerialLoader(
				yamlLoader,
				xload.OSLoader(), // OSLoader values take precedence over yaml
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

	rf, err := os.Open("request.default.yaml")
	if err != nil {
		panic(err)
	}

	reqYamlLoader, err := yaml.NewLoader(rf, "_")
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
				businessConfigLoader(), // businessConfigLoader values take precedence over yaml
			),
		),
	)
	if err != nil {
		panic(err)
	}

	prettyPrint(rcfg)

	_ = rf.Close()
	_ = f.Close()
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
