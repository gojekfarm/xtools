package main

import (
	"context"
	"time"

	"github.com/gojekfarm/xtools/xconfig"
)

func main() {
	ctx := context.Background()

	// application config is loaded once at startup
	// from yaml, and then from environment variables
	cfg := NewAppConfig()

	err := xconfig.Load(
		ctx,
		cfg,
		xconfig.Series(
			xconfig.YAMLLoader("application.yaml"),
			xconfig.OSLoader(),
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

	err = xconfig.Load(
		ctx,
		rcfg,
		xconfig.Series(
			xconfig.YAMLLoader("request.yaml"),
			BusinessConfigLoader(),
		),
	)
	if err != nil {
		panic(err)
	}
}

func NewAppConfig() *AppConfig {
	return &AppConfig{
		Name: "xconfig",
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
