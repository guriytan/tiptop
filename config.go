package tiptop

import (
	"errors"
	"time"
)

const (
	DefaultCleanWindows  = 30 * time.Minute
	DefaultShardSize     = 1024
	DefaultInitEntrySize = 5 * MB
)

// Config provides some environmental parameter to sustain tiptop running.
type Config struct {
	// Number of cache shard, used to be the power of 2.
	// Default of ShardSize is 1024.
	ShardSize int
	// Initialize size of entry in shard.
	InitEntrySize int
	// Max size of cache in Byte. if the use of cache in in-memory have achieved,
	// it will remove the entry by FIFO regulation.
	// Default value is set to 0 which mean unlimited size.
	MaxCacheSize int
	// CleanWindow is the period used to remove outdated entry.
	CleanWindow time.Duration
	// Expiration Time of the entry which is not assign.
	// DefaultTTL is set to 0 mean that entry never out of date.
	DefaultTTL time.Duration
	// tiptop use in-memory to caching acquiescently. When the Redis is on,
	// if the number of marker exceed the MaxEntrySize, the oldest entry will be remove to
	// Redis. if want to use Redis as secondary cache, set RedisAddr to be "addr:port"
	// The address required for connecting to Redis
	RedisAddr string
	// The password required for connecting to Redis
	// When RedisPwd is "" mean that Redis Client doesn't need password.
	RedisPwd string
	// RedisMinIdle, the minimum idle of the redis connection.
	RedisMinIdle int
	//RedisPoolSize, the pool size of the redis connection
	RedisPoolSize int
	// When the OnRemove is true, if the number of marker exceed the MaxEntrySize,
	// the oldest entry will be remove.
	OnRemove bool
}

func DefaultConfig() Config {
	return Config{
		ShardSize:     DefaultShardSize,
		CleanWindow:   DefaultCleanWindows,
		InitEntrySize: DefaultInitEntrySize,
		OnRemove:      true,
	}
}

func implementConfig(config *Config) error {
	if !isPowerOfTwo(config.ShardSize) {
		return errors.New("shard's size must be the power of 2")
	}
	if config.CleanWindow == 0 {
		config.CleanWindow = DefaultCleanWindows
	}
	if config.ShardSize == 0 {
		config.ShardSize = DefaultShardSize
	}
	return nil
}

// maximumShardSize computes maximum shard size
func (c Config) maximumShardSize() int {
	maxShardSize := 0

	if c.MaxCacheSize > 0 {
		maxShardSize = c.MaxCacheSize / c.ShardSize
	}

	return maxShardSize
}
