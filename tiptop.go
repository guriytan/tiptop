package tiptop

import (
	"time"
)

const (
	Byte = 1
	KB   = 1024 * Byte
	MB   = 1024 * KB
	GB   = 1024 * MB
)

// TipTop is the main entrance provided api to call by user.
type TipTop struct {
	shards    []*shard
	shardSize uint64
	hash      hashCalculator
	config    *Config
	close     chan bool
	shuffler  shuffler
}

// NewTipTop return a Tip-Top instance.
func NewTipTop(config Config) (*TipTop, error) {
	// check config
	err := implementConfig(&config)
	if err != nil {
		return nil, err
	}

	t := &TipTop{
		shards:    make([]*shard, config.ShardSize),
		shardSize: uint64(config.ShardSize - 1),
		hash:      defaultHashCalculator(),
		config:    &config,
		shuffler:  newDefaultShuffle(),
	}

	// init every shard
	for i := 0; i < config.ShardSize; i++ {
		t.shards[i] = initShard(&config)
	}

	// coroutines run
	t.tikTok()

	return t, nil
}

// tikTok run background to remove outdated entry.
func (t *TipTop) tikTok() {
	if t.config.CleanWindow > 0 {
		go func() {
			ticker := time.NewTicker(t.config.CleanWindow)
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					t.removeOutdated()
				case <-t.close:
					return
				}
			}
		}()
	}
}

// close is used to signal a shutdown of the cache when you are done with it.
func (t *TipTop) Close() error {
	close(t.close)
	for _, shard := range t.shards {
		shard.close()
	}
	return nil
}

// Get reads entry for the key.
func (t *TipTop) Get(key string) (value []byte, err error) {
	hash := t.hash.sum64(key)
	return t.getShard(hash).get(key, hash)
}

// Set saves entry under the key
func (t *TipTop) Set(key string, value []byte) error {
	return t.SetWithTTL(key, value, t.config.DefaultTTL)
}

// Set saves entry under the key with expiration
func (t *TipTop) SetWithTTL(key string, value []byte, ttl time.Duration) error {
	hash := t.hash.sum64(key)
	return t.getShard(hash).set(key, hash, value, ttl)
}

// Delete removes the key
func (t *TipTop) Delete(key string) error {
	hash := t.hash.sum64(key)
	return t.getShard(hash).del(hash)
}

// Reset empties all cache shards
func (t *TipTop) Reset() {
	for _, shard := range t.shards {
		shard.reset()
	}
}

func (t *TipTop) getShard(hash uint64) *shard {
	return t.shards[hash&t.shardSize]
}

func (t *TipTop) removeOutdated() {
	n := t.shuffler.shuffle(int(t.shardSize))
	for i := 0; i < n; i++ {
		t.shards[t.shuffler.shuffle(int(t.shardSize))].removeOutdated()
	}
}

// Len computes number of entries in cache
func (t *TipTop) Len() int {
	var l int
	for _, shard := range t.shards {
		l += shard.len()
	}
	return l
}

// Capacity returns amount of bytes store in the cache.
func (t *TipTop) Cap() int {
	var capacity int
	for _, shard := range t.shards {
		capacity += shard.cap()
	}
	return capacity
}

// Stats returns cache's statistics
func (t *TipTop) GetStats() Stats {
	var s Stats
	for _, shard := range t.shards {
		tmp := shard.getStats()
		s.Hits += tmp.Hits
		s.Misses += tmp.Misses
		s.HitsRedis += tmp.HitsRedis
		s.MissesRedis += tmp.MissesRedis
		s.Modify += tmp.Modify
		s.Collision += tmp.Collision
	}
	return s
}
