package tiptop

import (
	"bytes"
	"fmt"
	"log"
	"math/rand"
	"os"
	"runtime/pprof"
	"testing"
	"time"
)

func BenchmarkTipTop_Get(b *testing.B) {
	cpuProfile := "tip-top_get"
	f, err := os.Create(cpuProfile)
	if err != nil {
		log.Fatal(err)
	}
	_ = pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()
	for _, shards := range []int{512, 1024, 4096} {
		b.Run(fmt.Sprintf("%d-shards", shards), func(b *testing.B) {
			t, _ := NewTipTop(Config{
				ShardSize:    shards,
				MaxCacheSize: KB * shards,
				RedisAddr:    "172.18.29.81:6379",
				OnRemove:     true,
			})
			message := bytes.Repeat([]byte("a"), 2)
			for i := 0; i < b.N; i++ {
				_ = t.Set(fmt.Sprintf("key-%d", i), message)
			}

			b.ResetTimer()
			b.RunParallel(func(pb *testing.PB) {
				b.ReportAllocs()
				i := 0
				for pb.Next() {
					_, _ = t.Get(fmt.Sprintf("key-%d", i))
					i++
				}
			})
		})
	}
}

func BenchmarkTipTop_Set(b *testing.B) {
	cpuProfile := "tip-top_set"
	f, err := os.Create(cpuProfile)
	if err != nil {
		log.Fatal(err)
	}
	_ = pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	message := bytes.Repeat([]byte("a"), 256)
	for _, shards := range []int{512, 1024, 4096} {
		b.Run(fmt.Sprintf("%d-shards", shards), func(b *testing.B) {
			cache, _ := NewTipTop(Config{
				ShardSize:    shards,
				MaxCacheSize: KB * shards,
				OnRemove:     true,
			})
			rand.Seed(time.Now().Unix())

			b.RunParallel(func(pb *testing.PB) {
				id := rand.Int()
				counter := 0

				b.ReportAllocs()
				for pb.Next() {
					_ = cache.Set(fmt.Sprintf("key-%d-%d", id, counter), message)
					counter = counter + 1
				}
			})
		})
	}
}
