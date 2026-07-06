package main

import (
	"testing"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/konradreiche/cachebench"
)

func Benchmark(b *testing.B) {
	bench := cachebench.New()

	for workload := range bench.Workloads() {
		b.Run(workload.Name(), func(b *testing.B) {
			cache := newLRU(b, workload)
			b.ResetTimer()
			for range b.N {
				workload.Run(b, cache)
			}
		})
	}
}

type lruCache struct {
	cache *lru.Cache[string, string]
}

func newLRU(tb testing.TB, workload *cachebench.Workload) *lruCache {
	cache, err := lru.New[string, string](workload.CacheSize())
	if err != nil {
		tb.Fatal(err)
	}
	return &lruCache{
		cache: cache,
	}
}

func (l *lruCache) Load(key string) (string, bool) {
	return l.cache.Get(key)
}

func (l *lruCache) Store(key, value string) {
	l.cache.Add(key, value)
}
