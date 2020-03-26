package tiptop

// Stats is used to analyze the cache heat rate.
type Stats struct {
	// Hits is a number of successfully found keys
	Hits int64 `json:"hits"`
	// HitsRedis is a number of successfully found keys in redis
	HitsRedis int64 `json:"hits-redis"`
	// Misses is a number of not found keys
	Misses int64 `json:"misses"`
	// MissesRedis is a number of not found keys in redis
	MissesRedis int64 `json:"misses-redis"`
	// Collision is a number of happened key-collision
	Collision int64 `json:"collision"`
	// Modify is a number of happened key-modify
	Modify int64 `json:"stats-modify"`
	// Sync is a number of happened key sync from redis to in-memory
	Sync int64 `json:"stats-sync"`
}

func NewStats() Stats {
	return Stats{}
}
