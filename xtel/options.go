package xtel

import (
	"context"
	"time"

	"github.com/hashicorp/go-multierror"
)

const (
	newExporterTimeout = 15 * time.Second
)

// ProviderOption changes the behaviour of Provider.
type ProviderOption interface{ apply(*Provider) }

// DisableClientAutoTracing controls automatic tracing of downstream client connection.
// This is ENABLED by default.
var DisableClientAutoTracing = &disableClientTracing{}

// SamplingFraction configures sampling decision of traces. It makes a sampling decision based on the TraceID ratio.
// SamplingFraction >= 1 will always sample. SamplingFraction < 0 are treated as zero.
type SamplingFraction float64

func (sf SamplingFraction) apply(p *Provider) { p.samplingRatio = float64(sf) }

type disableClientTracing struct{}

func (d *disableClientTracing) apply(p *Provider) { p.roundTripper = nil }

func (fn TraceExporterFunc) apply(p *Provider) {
	ctx, cancel := context.WithTimeout(context.Background(), newExporterTimeout)
	defer cancel()

	se, err := fn(ctx)
	if se != nil {
		p.spanExporters = append(p.spanExporters, se)
	}

	p.initErrors = multierror.Append(p.initErrors, err)
}

func (fn MetricReaderFunc) apply(p *Provider) {
	ctx, cancel := context.WithTimeout(context.Background(), newExporterTimeout)
	defer cancel()

	me, err := fn(ctx)
	if me != nil {
		p.metricReaders = append(p.metricReaders, me)
	}

	p.initErrors = multierror.Append(p.initErrors, err)
}