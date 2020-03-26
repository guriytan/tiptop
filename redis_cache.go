package tiptop

import (
	"github.com/go-redis/redis"
	"strconv"
	"sync"
	"time"
)

type redisCache struct {
	client *redis.Client
}

const KeyPrefix = "tiptop::key::"

var (
	one   sync.Once
	cache *redisCache
)

func newRedisCache(config *Config) (*redisCache, error) {
	var err error
	one.Do(func() {
		client, err := newRedisClient(config)
		if err != nil {
			return
		}
		cache = &redisCache{client}
	})
	return cache, err
}

func (redis *redisCache) getKey(key uint64) ([]byte, error) {
	return redis.client.Get(KeyPrefix + strconv.FormatUint(key, 10)).Bytes()
}

func (redis *redisCache) setKey(key uint64, value []byte, expiration int64) {
	if expiration > 0 && expiration > time.Now().Unix() {
		redis.client.Set(KeyPrefix+strconv.FormatUint(key, 10), value, time.Duration(expiration-time.Now().Unix()))
	} else if expiration == 0 {
		redis.client.Set(KeyPrefix+strconv.FormatUint(key, 10), value, time.Duration(0))
	}
}

func (redis *redisCache) delKey(key uint64) {
	redis.client.Del(KeyPrefix + strconv.FormatUint(key, 10))
}

func (redis *redisCache) exist(key uint64) bool {
	return redis.client.Exists(KeyPrefix+strconv.FormatUint(key, 10)).Val() != 0
}

func (redis *redisCache) reset() {
	iterator := redis.client.Scan(0, KeyPrefix+"*", 10).Iterator()
	for iterator.Next() {
		redis.client.Del(iterator.Val())
	}
}

func (redis *redisCache) close() {
	_ = redis.client.Close()
}
