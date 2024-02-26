package exporter

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func Test_newUptimeCollector(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())

	reg := prometheus.NewRegistry()
	assert.NoError(t, reg.Register(newUptimeCollector(ctx)))

	time.Sleep(time.Second)

	assert.NoError(t, testutil.GatherAndCompare(reg, strings.NewReader(`# HELP system_wallclock The number of seconds passed since the node was started.
# TYPE system_wallclock counter
system_wallclock 1
`), "system_wallclock"))

	cancel()

	// allow collector to stop
	time.Sleep(100 * time.Millisecond)
}
