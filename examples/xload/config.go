package main

import "time"

// AppConfig is the application configuration
type AppConfig struct {
	Name string      `env:"NAME"`
	HTTP HTTPConfig  `env:",prefix=HTTP_"`
	GRPC *GRPCConfig `env:",prefix=GRPC_"`
	DB   DBConfig
}

// HTTPConfig is the HTTP server configuration
type HTTPConfig struct {
	Host  string `env:"HOST"`
	Port  int    `env:"PORT"`
	Debug bool   `env:"DEBUG"`
}

// GRPCConfig is the GRPC server configuration
type GRPCConfig struct {
	Host  string `env:"HOST"`
	Port  int    `env:"PORT"`
	Debug bool   `env:"DEBUG"`
}

// DBConfig is the database configuration
type DBConfig struct {
	URL string `env:"DB_URL"`
}

// RequestConfig is the per-request configuration
type RequestConfig struct {
	Timeout     time.Duration     `env:"TIMEOUT"`
	FeatureFlag bool              `env:"FEATURE_FLAG"`
	LaunchV3    *ExperimentConfig `env:",prefix=LAUNCH_V3_"`
}

// ExperimentConfig is the experiment configuration
type ExperimentConfig struct {
	Enabled      bool    `env:"ENABLED"`
	SearchRadius float64 `env:"SEARCH_RADIUS"`
	Cuisines     []Tag   `env:"CUISINES"`
}

// Tag is a tag
type Tag struct {
	Name string `json:"name"`
	Code string `json:"code"`
}
