package gocraft

import (
	"math/rand"
	"time"

	"github.com/gojek/work"
	"github.com/gojek/work/webui"
	"github.com/gomodule/redigo/redis"
	"github.com/rs/zerolog"

	"github.com/gojekfarm/xtools/xworker"
	"github.com/gojekfarm/xtools/xworker/gocraft/internal/sentinel"
)

var randInt63n = rand.New(rand.NewSource(time.Now().UnixNano())).Int63n

// Fulfiller is a wrapper over github.com/gojek/work to fulfil xworker.Adapter.
// It implements xworker.Enqueuer, xworker.PeriodicEnqueuer and xworker.Registerer.
type Fulfiller struct {
	namespace            string
	logger               zerolog.Logger
	pool                 *work.WorkerPool
	client               *work.Client
	enqueuer             enqueuer
	webuiServer          *webui.Server
	disableCodec         bool
	useRawEncodedPayload bool
	collectFailOverFunc  func(string, time.Duration)
}

var _ xworker.Fulfiller = (*Fulfiller)(nil)

// New creates a Fulfiller for xworker.Adapter with provided Options.
func New(opts Options) *Fulfiller {
	if opts.Namespace == "" {
		opts.Namespace = "goWorker"
	}

	if opts.MaxConcurrency == 0 {
		opts.MaxConcurrency = 25
	}

	return &Fulfiller{
		namespace:            opts.Namespace,
		logger:               opts.Logger,
		disableCodec:         opts.DisableCodec,
		useRawEncodedPayload: opts.UseRawEncodedPayload,
		pool:                 work.NewWorkerPool(struct{}{}, uint(opts.MaxConcurrency), opts.Namespace, opts.Pool),
		client:               work.NewClient(opts.Namespace, opts.Pool),
		enqueuer:             work.NewEnqueuer(opts.Namespace, opts.Pool),
		webuiServer:          webui.NewServer(opts.Namespace, opts.Pool, opts.UIServerListenAddr),
	}
}

// NewWithSentinel creates a Fulfiller for xworker.Adapter with provided Options and SentinelPoolOptions.
func NewWithSentinel(opts Options, sentinelOpts SentinelPoolOptions) *Fulfiller {
	sentinelOpts.setDefaults()

	s := &sentinel.Sentinel{
		Addrs:      sentinelOpts.Addresses,
		MasterName: sentinelOpts.MasterName,
		Dial: func(addr string) (redis.Conn, error) {
			return redis.Dial("tcp",
				addr,
				redis.DialConnectTimeout(sentinelOpts.DialTimeout),
				redis.DialReadTimeout(sentinelOpts.ReadTimeout))
		},
	}

	if opts.Pool == nil {
		opts.Pool = &redis.Pool{
			MaxIdle:     sentinelOpts.MaxIdleConnections,
			MaxActive:   sentinelOpts.MaxActiveConnections,
			Wait:        true,
			IdleTimeout: sentinelOpts.IdleConnectionTimeout,
			Dial:        dialSentinelRedisWithRetry(sentinelOpts, s),
		}
	}

	f := New(opts)

	if sentinelOpts.RestartWorkerPoolOnFailOver {
		s.RegisterFailoverCallback(f.sentinelFailOverCallback)
	}

	return f
}

// Start starts the Fulfiller and associated processes.
func (f *Fulfiller) Start() error {
	f.pool.Start()

	if f.webuiServer != nil {
		f.webuiServer.Start()
	}

	return nil
}

// Stop stops the Fulfiller and associated processes.
func (f *Fulfiller) Stop() error {
	f.pool.Stop()

	if f.webuiServer != nil {
		f.webuiServer.Stop()
	}

	return nil
}

// WebUIServer returns a webui.Server with the configured Options.Pool.
func (f *Fulfiller) WebUIServer() *webui.Server {
	return f.webuiServer
}

func (f *Fulfiller) sentinelFailOverCallback(oldMasterAddr, newMasterAddr string) {
	if !f.pool.Started() {
		return
	}

	jitter := time.Duration(randInt63n(10000)) * time.Millisecond
	time.Sleep(jitter)

	start := time.Now()

	if f.collectFailOverFunc != nil {
		defer func() {
			f.collectFailOverFunc(f.namespace, time.Since(start))
		}()
	}

	f.pool.Stop()
	f.pool.Start()

	f.logger.Info().Fields(map[string]interface{}{
		"oldMaster": oldMasterAddr,
		"newMaster": newMasterAddr,
		"took":      time.Since(start).String(),
		"jitter":    jitter.String(),
	}).Msg("Restarted worker pool")
}
