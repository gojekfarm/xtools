package viper

import (
	"context"
	"strings"

	"github.com/spf13/viper"

	"github.com/gojekfarm/xtools/xload"
)

// Option allows configuring the Loader.
type Option interface{ apply(*options) }

type optionFunc func(*options)

func (f optionFunc) apply(o *options) { f(o) }

// From allows passing a pre-configured Viper instance.
func From(v *viper.Viper) Option { return optionFunc(func(o *options) { o.viper = v }) }

// ValueMapper allows specifying a custom value mapper function that will be used to flatten the config
// for xload.Loader from Viper.
type ValueMapper func(map[string]any) map[string]string

func (m ValueMapper) apply(o *options) { o.mapValues = m }

// ConfigFile allows specifying the config file to be used.
type ConfigFile string

func (p ConfigFile) apply(o *options) { o.file.absPath = string(p) }

// ConfigName allows specifying the config file name.
type ConfigName string

func (n ConfigName) apply(o *options) { o.file.name = string(n) }

// ConfigType allows specifying the config file type.
type ConfigType string

func (t ConfigType) apply(o *options) { o.file.ext = string(t) }

// ConfigPaths allows specifying the config file search paths.
type ConfigPaths []string

// AutoEnv allows enabling/disabling automatic environment variable loading.
type AutoEnv bool

func (b AutoEnv) apply(o *options) { o.autoEnv = bool(b) }

func (p ConfigPaths) apply(o *options) {
	for _, path := range p {
		o.file.paths = append(o.file.paths, path)
	}
}

// Transformer allows specifying a custom transformer function.
type Transformer func(v *viper.Viper, next xload.Loader) xload.Loader

func (t Transformer) apply(o *options) { o.transform = t }

type options struct {
	viper     *viper.Viper
	file      fileOpts
	transform Transformer
	mapValues ValueMapper
	autoEnv   bool
}

type fileOpts struct {
	absPath string
	name    string
	ext     string
	paths   []string
}

func def() *options {
	return &options{
		viper: viper.New(),
		file: fileOpts{
			name:  "application",
			ext:   "yaml",
			paths: []string{"./", "../"},
		},
		autoEnv:   true,
		mapValues: func(in map[string]any) map[string]string { return xload.FlattenMap(in, "_") },
		transform: func(_ *viper.Viper, next xload.Loader) xload.Loader {
			return xload.LoaderFunc(func(ctx context.Context, key string) (string, error) {
				return next.Load(ctx, strings.ToLower(key))
			})
		},
	}
}
