package tiptop

import "math/rand"

type shuffler interface {
	shuffle(border int) int
}

type defaultShuffler struct {
	r *rand.Rand
}

func newDefaultShuffle() shuffler {
	return &defaultShuffler{}
}

func (s *defaultShuffler) shuffle(border int) int {
	return s.r.Intn(border)
}
