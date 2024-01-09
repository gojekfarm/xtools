package viper

import (
	"context"
	"errors"

	"github.com/spf13/viper"

	"github.com/gojekfarm/xtools/xload"
)

// Loader is a xload.Loader implementation that uses viper.
type Loader struct {
	v  *viper.Viper
	kl xload.Loader
}

// New creates a new Loader instance.
func New(options ...Option) (*Loader, error) {
	opts := def()

	for _, o := range options {
		o.apply(opts)
	}

	v := opts.viper

	if opts.file.absPath != "" {
		v.SetConfigFile(opts.file.absPath)
	} else {
		v.SetConfigName(opts.file.name)
		v.SetConfigType(opts.file.ext)

		for _, path := range opts.file.paths {
			v.AddConfigPath(path)
		}
	}

	if opts.autoEnv {
		v.AutomaticEnv()
	}

	var notFoundErr viper.ConfigFileNotFoundError
	if err := v.ReadInConfig(); !errors.As(err, &notFoundErr) && err != nil {
		return nil, err
	}

	mp := make(map[string]string)
	for key, value := range xload.FlattenMap(v.AllSettings(), opts.separator) {
		mp[key] = value
	}

	return &Loader{
		v: v,
		kl: opts.transform(v, xload.LoaderFunc(func(_ context.Context, key string) (string, error) {
			return mp[key], nil
		})),
	}, nil
}

// ConfigFileUsed returns the config file used, if any.
func (l *Loader) ConfigFileUsed() string { return l.v.ConfigFileUsed() }

// Load returns the value for the given key.
func (l *Loader) Load(ctx context.Context, key string) (string, error) {
	return l.kl.Load(ctx, key)
}
