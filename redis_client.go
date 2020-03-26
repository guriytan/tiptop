package tiptop

import (
	"github.com/go-redis/redis"
	"time"
)

const (
	MinIdle      = 5
	PoolSize     = 20
	RedisTimeout = 3000
)

func newRedisClient(config *Config) (*redis.Client, error) {
	minIdle := MinIdle
	if config.RedisMinIdle != 0 {
		minIdle = config.RedisMinIdle
	}
	poolSize := PoolSize
	if config.RedisPoolSize != 0 {
		poolSize = config.RedisPoolSize
	}
	client := redis.NewClient(&redis.Options{
		Addr:         config.RedisAddr,
		Password:     config.RedisPwd,
		MinIdleConns: minIdle,
		PoolSize:     poolSize,
		IdleTimeout:  time.Duration(RedisTimeout),
	})
	if err := client.Ping().Err(); err != nil {
		return nil, err
	}
	return client, nil
}
