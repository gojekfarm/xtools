package gocraft

import (
	"time"

	"github.com/gomodule/redigo/redis"
	"github.com/rs/zerolog"
)

// Options holds the options for Fulfiller.
type Options struct {
	Namespace      string
	Pool           *redis.Pool
	Logger         zerolog.Logger
	MaxConcurrency int
	// DEPRECATED: DisableCodec can be used to opt out of Payload Codec functionality of xworker.
	DisableCodec bool
	// When UseRawEncodedPayload is true, no additional operation is done on the encoded payload.
	// Thus, for gocraft either the encoded payload needs to be json supported data type, or the encoder &
	// decoder needs to ensure the encoded payload is encapsulated in string.
	// e.g. for an encoded payload 0jvwb2ErhYuOmIdx6M0qPI"REVMWOhXS5, the encoder need to modify it to
	// "0jvwb2ErhYuOmIdx6M0qPI\"REVMWOhXS5" to ensure the payload is still a valid json supported data type.
	// This should not be an issue for default encoder given it returns valid json bytes.
	UseRawEncodedPayload bool
	UIServerListenAddr   string
}

// SentinelPoolOptions is the configuration used to create a connection
// pool. MaxDialAttempts retries are attempted for dialling a connection
// with constant backoff of DialBackoffDelay.
type SentinelPoolOptions struct {
	Addresses                   []string
	MasterName                  string
	DialTimeout                 time.Duration
	ReadTimeout                 time.Duration
	WriteTimeout                time.Duration
	MaxDialAttempts             int
	DialBackoffDelay            time.Duration
	MaxActiveConnections        int
	MaxIdleConnections          int
	IdleConnectionTimeout       time.Duration
	RestartWorkerPoolOnFailOver bool
	RedisPassword               string
}

func (s *SentinelPoolOptions) setDefaults() {
	if s.MasterName == "" {
		s.MasterName = "mymaster"
	}
}
