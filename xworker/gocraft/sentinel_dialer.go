package gocraft

import (
	"context"

	"github.com/gomodule/redigo/redis"
	"github.com/sethvargo/go-retry"

	"github.com/gojekfarm/xtools/xworker/gocraft/internal/sentinel"
)

func dialSentinelRedisWithRetry(opts SentinelPoolOptions, s *sentinel.Sentinel) func() (conn redis.Conn, err error) {
	return func() (conn redis.Conn, err error) {
		d := func(context.Context) error {
			masterAddr, err := s.MasterAddr()
			if err != nil {
				return err
			}

			conn, err = redis.Dial("tcp", masterAddr,
				redis.DialConnectTimeout(opts.DialTimeout),
				redis.DialReadTimeout(opts.ReadTimeout),
				redis.DialWriteTimeout(opts.WriteTimeout),
				redis.DialPassword(opts.RedisPassword),
			)

			return retry.RetryableError(err)
		}

		bs, err := retry.NewConstant(opts.DialBackoffDelay)
		if err != nil {
			return nil, err
		}

		if opts.MaxDialAttempts > 0 {
			bs = retry.WithMaxRetries(uint64(opts.MaxDialAttempts), bs)
		}

		err = retry.Do(context.Background(), bs, d)

		return
	}
}
