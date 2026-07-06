// Package cachebench provides an API for running workloads to benchmark cache
// implementations.
package cachebench

import (
	"iter"
	"testing"

	"github.com/konradreiche/cachebench/internal/generator"
)

type Benchmark struct {
	gen *generator.FiniteZipf
}

func New(tb testing.TB) *Benchmark {
	gen, err := generator.NewFiniteZipf(500_000, 0.75, generator.WithSeed(1))
	if err != nil {
		tb.Fatal(err)
	}
	return &Benchmark{
		gen: gen,
	}
}

func (b *Benchmark) Workloads() iter.Seq[*Workload] {
	return b.zipfWorkloads()
}

func (b *Benchmark) zipfWorkloads() iter.Seq[*Workload] {
	w := &Workload{
		cacheSize: 50_000,
		keyspace:  500_000,
		gen:       b.gen,
		stats:     &stats{},
	}
	return func(yield func(*Workload) bool) {
		yield(w)
	}
}
