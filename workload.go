package cachebench

import (
	"sync/atomic"
	"testing"
	"time"

	"github.com/konradreiche/cachebench/internal/generator"
)

type Workload struct {
	cacheSize int
	keyspace  uint64
	gen       *generator.FiniteZipf
	stats     []*stats
}

func (w *Workload) RunParallel(b *testing.B, cache Cache) {
	var i atomic.Uint64
	b.RunParallel(func(pb *testing.PB) {
		id := i.Add(1) - 1
		for pb.Next() {
			w.Process(b, cache, id)
		}
	})
	w.RecordMetrics(b)
}

type Cache interface {
	Store(key, value string)
	Load(key string) (string, bool)
}

func (w *Workload) Process(b *testing.B, cache Cache, id uint64) {
	key := w.Next()
	start := time.Now()
	_, ok := cache.Load(key)
	w.stats[id].recordLoadDuration(start)
	if !ok {
		w.stats[id].misses++
		start = time.Now()
		cache.Store(key, key)
		w.stats[id].recordStoreDuration(start)
		return
	}
	w.stats[id].hits++
}

func (w *Workload) RecordMetrics(b *testing.B) {
	stats := newStats()
	for _, stat := range w.stats {
		stats.hits += stat.hits
		stats.misses += stat.misses
		stats.totalLoadTime += stat.totalStoreTime
		stats.totalStoreTime += stat.totalStoreTime
	}
	b.ReportMetric(stats.hitRate()*100, "hitRate")
	b.ReportMetric(stats.loadDuration(), "load-ns/op")
	b.ReportMetric(stats.storeDuration(), "store-ns/op")
}

func (w *Workload) Next() string {
	return w.gen.Next()
}

func (w *Workload) CacheSize() int {
	return w.cacheSize
}

func (w *Workload) Name() string {
	return "zipf"
}

type stats struct {
	hits   int
	misses int

	totalLoadTime  time.Duration
	totalStoreTime time.Duration
}

func newStats() *stats {
	return &stats{}
}

func (s *stats) recordLoadDuration(start time.Time) {
	s.totalLoadTime += time.Since(start)
}

func (s *stats) recordStoreDuration(start time.Time) {
	s.totalStoreTime += time.Since(start)
}

func (s *stats) loadDuration() float64 {
	total := float64(s.hits + s.misses)
	return float64(s.totalLoadTime) / total
}

func (s *stats) storeDuration() float64 {
	return float64(s.totalStoreTime) / float64(s.misses)
}

func (s *stats) hitRate() float64 {
	total := float64(s.hits + s.misses)
	return float64(s.hits) / total
}
