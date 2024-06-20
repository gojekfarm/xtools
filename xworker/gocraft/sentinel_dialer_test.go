package gocraft

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/gojekfarm/xtools/xworker/gocraft/internal/sentinel"
)

func Test_dialSentinelRedisWithRetry(t *testing.T) {
	f := dialSentinelRedisWithRetry(SentinelPoolOptions{}, &sentinel.Sentinel{})

	_, err := f()
	assert.EqualError(t, err, "t must be greater than 0")
}
