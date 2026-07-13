package main

import (
	"testing"

	lru "github.com/hashicorp/golang-lru/v2"
	"github.com/konradreiche/cachebench"
	"github.com/konradreiche/cachebench/benchmark/internal/lfu"
)

func Benchmark(b *testing.B) {
	benchmarks := []benchmark{
		{
			name: "lru",
			newCache: func(b *testing.B, workload *cachebench.Workload) cachebench.Cache {
				return newLRU(b, workload)
			},
		},
		{
			name: "lfu",
			newCache: func(b *testing.B, w *cachebench.Workload) cachebench.Cache {
				cache, err := lfu.New[string, string](w.CacheSize())
				if err != nil {
					b.Fatal(err)
				}
				return cache
			},
		},
	}

	for _, bb := range benchmarks {
		b.Run(bb.name, func(b *testing.B) {
			bench := cachebench.New(b, cachebench.WithParallelism(1))
			for workload := range bench.Workloads() {
				cache := bb.newCache(b, workload)
				workload.RunParallel(b, cache)
			}
		})
	}
}

type benchmark struct {
	newCache func(b *testing.B, w *cachebench.Workload) cachebench.Cache
	name     string
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
