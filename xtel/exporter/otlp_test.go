package exporter

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"

	"github.com/gojekfarm/xtools/xtel"
)

func TestNewOTLP(t *testing.T) {
	tests := []struct {
		name   string
		option []Option
	}{
		{
			name:   "Default",
			option: []Option{WithTracesExporterInsecure},
		},
		{
			name:   "HTTP",
			option: []Option{HTTP, WithTracesExporterInsecure},
		},
		{
			name:   "Without Insecure Transport Credentials",
			option: []Option{HTTP},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			p, err := xtel.NewProvider("test", NewOTLP(tc.option...))
			assert.NoError(t, err)
			assert.NotNil(t, p)
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := newConfig()
	expected := &config{exporterEndpoint: defaultExporterEndpoint, headers: map[string]string{}}

	assert.Equal(t, expected, cfg)
}

func TestNewTraceExporter(t *testing.T) {
	newTraceExporter = func(ctx context.Context, client otlptrace.Client) (*otlptrace.Exporter, error) {
		return nil, errors.New("context already closed")
	}
	defer func() { newTraceExporter = otlptrace.New }()

	p, err := xtel.NewProvider("test-service", NewOTLP())
	assert.Error(t, err)
	assert.Nil(t, p)
}

func TestConfigurationOverrides(t *testing.T) {
	conf := newConfig(
		WithTracesExporterEndpoint("override-satellite-url"),
		WithTracesExporterInsecure,
	)
	emptyconfig := newConfig()
	assert.NotNil(t, emptyconfig)
	expected := &config{
		exporterEndpoint:               "override-satellite-url",
		tracesExporterEndpointInsecure: true,
		headers:                        map[string]string{},
	}
	assert.Equal(t, expected, conf)
}
