package main

import "time"

type AppConfig struct {
	Name string      `env:"NAME"`
	HTTP HTTPConfig  `env:",prefix=HTTP_"`
	GRPC *GRPCConfig `env:",prefix=GRPC_"`
	DB   DBConfig
}

type HTTPConfig struct {
	Host  string `env:"HOST"`
	Port  int    `env:"PORT"`
	Debug bool   `env:"DEBUG"`
}

type GRPCConfig struct {
	Host  string `env:"HOST"`
	Port  int    `env:"PORT"`
	Debug bool   `env:"DEBUG"`
}

type DBConfig struct {
	URL string `env:"DB_URL"`
}

type RequestConfig struct {
	Timeout     time.Duration     `env:"TIMEOUT"`
	FeatureFlag bool              `env:"FEATURE_FLAG"`
	LaunchV3    *ExperimentConfig `env:",prefix=LAUNCH_V3_"`
}

type ExperimentConfig struct {
	Enabled      bool    `env:"ENABLED"`
	SearchRadius float64 `env:"SEARCH_RADIUS"`
	Cuisines     []Tag   `env:"CUISINES"`
}

type Tag struct {
	Name string `env:"NAME"`
	Code string `env:"CODE"`
}
