package main

import (
	"context"
	"time"

	"github.com/gojekfarm/xtools/xload"
)

func BusinessConfigLoader() xload.Loader {
	return xload.LoaderFunc(func(ctx context.Context, key string) (string, error) {
		// fetch from remote config
		return "", nil
	})
}

func YAMLLoader(filename string) xload.Loader {
	// read into a flattened map
	// return a MapLoader
	return xload.MapLoader{}
}

func main() {
	ctx := context.Background()

	// application config is loaded once at startup
	// from yaml, and then from environment variables
	cfg := NewAppConfig()

	err := xload.Load(
		ctx,
		cfg,
		xload.WithLoader(
			xload.SerialLoader(
				YAMLLoader("application.yaml"),
				xload.OSLoader(),
			),
		),
	)
	if err != nil {
		panic(err)
	}

	// request config is loaded for every request
	// defaults are set in the code, and then
	// overridden by remote config.
	//
	// This can be used in HTTP, GRPC, handlers or
	// middleware.
	rcfg := NewRequestConfig()

	err = xload.Load(
		ctx,
		rcfg,
		xload.WithLoader(
			xload.SerialLoader(
				YAMLLoader("request.yaml"),
				BusinessConfigLoader(),
			),
		),
	)
	if err != nil {
		panic(err)
	}
}

func NewAppConfig() *AppConfig {
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

func NewRequestConfig() *RequestConfig {
	return &RequestConfig{
		Timeout:     5 * time.Second,
		FeatureFlag: true,
		LaunchV3: &ExperimentConfig{
			Enabled:      false,
			SearchRadius: 2.5,
		},
	}
}
