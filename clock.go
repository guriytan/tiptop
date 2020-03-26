package tiptop

import "time"

type clock interface {
	epoch() int64
	exp(ttl time.Duration) int64
}

func newDefaultClock() clock {
	return defaultClock{}
}

type defaultClock struct {
}

func (c defaultClock) epoch() int64 {
	return time.Now().Unix()
}

func (c defaultClock) exp(ttl time.Duration) int64 {
	if ttl != 0 {
		return time.Now().Add(ttl).Unix()
	}
	return 0
}
