package cachebench

import (
	"iter"
	"testing"

	"github.com/konradreiche/cachebench/internal/generator"
)

type Workload struct {
	cacheSize int
	keyspace  uint64
	gen       *generator.FiniteZipf
	stats     *stats
}

func (w *Workload) RecordMetrics(b *testing.B) {
	b.ReportMetric(w.stats.hitRate()*100, "hitRate")
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
}

func (s *stats) hitRate() float64 {
	total := float64(s.hits + s.misses)
	return float64(s.hits) / total
}
