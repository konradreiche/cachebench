// Package cachebench provides an API for running workloads to benchmark cache
// implementations.
package cachebench

import (
	"iter"

	"github.com/konradreiche/cachebench/internal/generator"
)

type Benchmark struct{}

func New() *Benchmark {
	return &Benchmark{}
}

func (b *Benchmark) Workloads() iter.Seq[*Workload] {
	return b.zipfWorkloads()
}

func (b *Benchmark) zipfWorkloads() iter.Seq[*Workload] {
	w := &Workload{
		cacheSize: 50_000,
		keyspace:  500_000,
		gen:       generator.NewFiniteZipf(500_000, 0.75),
		stats:     &stats{},
	}
	return func(yield func(*Workload) bool) {
		yield(w)
	}
}
