package gocraft

import (
	"testing"

	"github.com/gojekfarm/xtools/xworker/gocraft/internal/sentinel"
	"github.com/stretchr/testify/assert"
)

func Test_dialSentinelRedisWithRetry(t *testing.T) {
	f := dialSentinelRedisWithRetry(SentinelPoolOptions{}, &sentinel.Sentinel{})

	_, err := f()
	assert.EqualError(t, err, "t must be greater than 0")
}
