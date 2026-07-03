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
			cache, err := lru.New[string, string](workload.CacheSize())
			if err != nil {
				b.Fatal(err)
			}
			b.ResetTimer()
			for range b.N {
				for key := range workload.Next() {
					if _, ok := cache.Get(key); !ok {
						workload.RecordMiss()
						cache.Add(key, key)
						continue
					}
					workload.RecordHit()
				}
				workload.RecordMetrics(b)
			}
		})
	}
}
