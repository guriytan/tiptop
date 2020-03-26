package tiptop

import (
	"errors"
	"hash/crc32"
	"sync"
	"sync/atomic"
	"time"
)

// shard is an unit of the in-memory cache which use map to store data.
// every key-value will be calculated that which shard is chosen to store it.
type shard struct {
	lock    sync.RWMutex
	marker  map[uint64]int
	entries ByteQueue
	buffer  []byte

	redisCache    *redisCache
	redisEnable   bool
	onRemove      bool
	InitEntrySize int

	clock    clock
	stats    Stats
	shuffler shuffler
}

var (
	errKeyNotFound = errors.New("key is not found")
	errEntryIsDead = errors.New("key is outdated")
	errMaxEntry    = errors.New("entry is bigger than max shard size")
)

func initShard(config *Config) *shard {
	shard := &shard{
		marker:   make(map[uint64]int),
		entries:  NewByteQueue(config.InitEntrySize, config.maximumShardSize()),
		buffer:   make([]byte, config.InitEntrySize),
		lock:     sync.RWMutex{},
		onRemove: config.OnRemove,

		clock:         newDefaultClock(),
		shuffler:      newDefaultShuffle(),
		InitEntrySize: config.InitEntrySize,
	}
	if config.OnRemove && config.RedisAddr != "" {
		client, err := newRedisCache(config)
		if err != nil {
			panic("init shard err: " + err.Error())
		}

		shard.redisCache = client
		shard.redisEnable = true
	}
	return shard
}

// getWrappedEntry get the entry by the index of the hashmap
// if the key doesn't exist in hashmap and the RedisEnable is true,
// entry will search from the redis, and sync to the in-memory.
func (s *shard) getWrappedEntry(hash uint64) ([]byte, error) {
	s.lock.RLock()
	itemIndex := s.marker[hash]

	if itemIndex == 0 {
		s.lock.RUnlock()
		s.statsMiss()
		return nil, errKeyNotFound
	}

	wrappedEntry, err := s.entries.Get(itemIndex)
	if err != nil {
		if s.redisEnable {
			bytes, err := s.redisCache.getKey(hash)

			s.lock.RUnlock()
			if err != nil {
				s.statsMissRedis()
				return nil, errKeyNotFound
			}

			go s.sync(hash, bytes)
			s.statsHitRedis()
			return bytes, nil
		}
		s.lock.RUnlock()
		s.statsMiss()
		return nil, err
	}
	s.lock.RUnlock()
	return wrappedEntry, nil
}

// get by read the entry from the entries queue.
// the crc32 will be checked to ensure the collision doesn't happened.
// the expiration time also will be checked. If the key is outdated, errEntryIsDead will be returned.
func (s *shard) get(key string, hash uint64) ([]byte, error) {
	wrappedEntry, err := s.getWrappedEntry(hash)
	if err != nil {
		return nil, err
	}

	s.lock.RLock()
	if crc := readCRC32FromEntry(wrappedEntry); crc != crc32.ChecksumIEEE([]byte(key)) {
		s.lock.RUnlock()
		s.statsCollision()
		return nil, errKeyNotFound
	}

	timeStamp := readTimestampFromEntry(wrappedEntry)
	if timeStamp != 0 && s.clock.epoch() > timeStamp {
		s.lock.RUnlock()
		go s.del(hash)
		return nil, errEntryIsDead
	}

	entry := readEntry(wrappedEntry)
	s.lock.RUnlock()
	s.statsHit()
	return entry, nil
}

// sync is a synchronization to keep the data read from redis store to in-memory
func (s *shard) sync(hash uint64, value []byte) {
	s.lock.Lock()
	if previousIndex := s.marker[hash]; previousIndex != 0 {
		s.lock.Unlock()
		return
	}
	for {
		if index, err := s.entries.Push(value); err == nil {
			s.marker[hash] = index
			s.redisCache.delKey(hash)
			s.lock.Unlock()
			s.statsSync()
			return
		}
		if !s.onRemove || s.removeOldest() != nil {
			s.lock.Unlock()
			return
		}
	}
}

func (s *shard) set(key string, hash uint64, value []byte, ttl time.Duration) error {
	s.lock.Lock()

	if previousIndex := s.marker[hash]; previousIndex != 0 {
		if previousEntry, err := s.entries.Get(previousIndex); err == nil {
			resetKeyFromEntry(previousEntry)
		}
	}

	w := wrapEntry(s.clock.exp(ttl), hash, crc32.ChecksumIEEE([]byte(key)), value, &s.buffer)

	for {
		if index, err := s.entries.Push(w); err == nil {
			s.marker[hash] = index
			s.lock.Unlock()
			s.statsModify()
			return nil
		}
		if !s.onRemove || s.removeOldest() != nil {
			s.lock.Unlock()
			return errMaxEntry
		}
	}
}

// del the key from hashmap , entries and redis if the key exist in redis,
func (s *shard) del(hash uint64) error {
	s.statsModify()

	// pre-check the key
	s.lock.RLock()
	itemIndex := s.marker[hash]
	if itemIndex == 0 {
		if s.redisEnable {
			if !s.redisCache.exist(hash) {
				s.lock.RUnlock()
				return errKeyNotFound
			}
		} else {
			s.lock.RUnlock()
			return errKeyNotFound
		}
	} else if err := s.entries.CheckGet(itemIndex); err != nil {
		s.lock.RUnlock()
		return err
	}
	s.lock.RUnlock()

	s.lock.Lock()
	itemIndex = s.marker[hash]

	if itemIndex == 0 {
		if s.redisEnable {
			if s.redisCache.exist(hash) {
				s.redisCache.delKey(hash)
				s.lock.Unlock()
				return nil
			}
		}
		s.lock.Unlock()
		return errKeyNotFound
	}

	wrappedEntry, err := s.entries.Get(itemIndex)
	if err != nil {
		s.lock.Unlock()
		return err
	}

	delete(s.marker, hash)
	resetKeyFromEntry(wrappedEntry)
	s.lock.Unlock()
	return nil
}

// remove outdated entry periodically
func (s *shard) removeOutdated() {
	r := s.shuffler.shuffle(len(s.marker))
	s.lock.Lock()
	for k, v := range s.marker {
		if r == 0 {
			break
		}

		if v == 0 {
			continue
		}
		wrappedEntry, err := s.entries.Get(v)
		if err != nil {
			continue
		}
		timeStamp := readTimestampFromEntry(wrappedEntry)
		if timeStamp != 0 && s.clock.epoch() > timeStamp {
			delete(s.marker, k)
			resetKeyFromEntry(wrappedEntry)
		}
		r--
	}
	s.lock.Unlock()
}

// removeOldest remove the oldest entry by popping the first entry
func (s *shard) removeOldest() error {
	oldest, err := s.entries.Pop()
	if err == nil {
		hash := readHashFromEntry(oldest)
		delete(s.marker, hash)
		if s.redisEnable {
			s.redisCache.setKey(hash, oldest, readTimestampFromEntry(oldest))
		}
		return nil
	}
	return err
}

func (s *shard) reset() {
	s.lock.Lock()
	defer s.lock.Unlock()

	s.marker = make(map[uint64]int)
	s.buffer = make([]byte, s.InitEntrySize)

	s.stats = NewStats()
	s.entries.Reset()

	if s.redisEnable {
		s.redisCache.reset()
	}
}

func (s *shard) close() {
	if s.redisEnable {
		s.redisCache.close()
	}
}

func (s *shard) len() int {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return len(s.marker)
}

func (s *shard) cap() int {
	s.lock.RLock()
	defer s.lock.RUnlock()

	return s.entries.Capacity()
}

func (s *shard) statsHit() {
	atomic.AddInt64(&s.stats.Hits, 1)
}

func (s *shard) statsMiss() {
	atomic.AddInt64(&s.stats.Misses, 1)
}

func (s *shard) statsHitRedis() {
	atomic.AddInt64(&s.stats.HitsRedis, 1)
}

func (s *shard) statsMissRedis() {
	atomic.AddInt64(&s.stats.MissesRedis, 1)
}

func (s *shard) statsCollision() {
	atomic.AddInt64(&s.stats.Collision, 1)
}

func (s *shard) statsModify() {
	atomic.AddInt64(&s.stats.Modify, 1)
}

func (s *shard) statsSync() {
	atomic.AddInt64(&s.stats.Sync, 1)
}

func (s *shard) getStats() Stats {
	return Stats{
		Hits:        atomic.LoadInt64(&s.stats.Hits),
		Misses:      atomic.LoadInt64(&s.stats.Misses),
		HitsRedis:   atomic.LoadInt64(&s.stats.HitsRedis),
		MissesRedis: atomic.LoadInt64(&s.stats.MissesRedis),
		Modify:      atomic.LoadInt64(&s.stats.Modify),
		Collision:   atomic.LoadInt64(&s.stats.Collision),
		Sync:        atomic.LoadInt64(&s.stats.Sync),
	}
}
