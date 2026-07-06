package cachebench

import (
	"iter"
	"testing"
	"time"

	"github.com/konradreiche/cachebench/internal/generator"
)

type Workload struct {
	cacheSize int
	keyspace  uint64
	gen       *generator.FiniteZipf
	stats     *stats
}

type cache interface {
	Store(key, value string)
	Load(key string) (string, bool)
}

func (w *Workload) Run(b *testing.B, cache cache) {
	for key := range w.Next() {
		start := time.Now()
		_, ok := cache.Load(key)
		w.stats.recordLoadDuration(start)
		if !ok {
			w.RecordMiss()
			start = time.Now()
			cache.Store(key, key)
			w.stats.recordStoreDuration(start)
			continue
		}
		w.RecordHit()
	}
	w.RecordMetrics(b)
}

func (w *Workload) RecordMetrics(b *testing.B) {
	b.ReportMetric(w.stats.hitRate()*100, "hitRate")
	b.ReportMetric(w.stats.loadDuration(), "load-ns/op")
	b.ReportMetric(w.stats.storeDuration(), "store-ns/op")
}

func (w *Workload) Next() iter.Seq[string] {
	return w.gen.Next()
}

func (w *Workload) RecordHit() {
	w.stats.hits++
}

func (w *Workload) RecordMiss() {
	w.stats.misses++
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
