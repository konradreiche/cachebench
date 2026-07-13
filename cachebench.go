// Package cachebench provides an API for running workloads to benchmark cache
// implementations.
package cachebench

import (
	"iter"
	"runtime"
	"testing"

	"github.com/konradreiche/cachebench/internal/generator"
)

type Benchmark struct {
	gen      *generator.FiniteZipf
	numProcs int
}

func New(b *testing.B, opts ...Option) *Benchmark {
	cfg := options{
		parallelism: 1,
	}
	if err := WithOptions(opts...)(&cfg); err != nil {
		b.Fatal(err)
	}
	gen, err := generator.NewFiniteZipf(500_000, 0.75, generator.WithSeed(1))
	if err != nil {
		b.Fatal(err)
	}
	b.ReportAllocs()
	b.SetParallelism(cfg.parallelism)
	return &Benchmark{
		gen:      gen,
		numProcs: runtime.GOMAXPROCS(0) * cfg.parallelism,
	}
}

func (b *Benchmark) Workloads() iter.Seq[*Workload] {
	return b.zipfWorkloads()
}

func (b *Benchmark) zipfWorkloads() iter.Seq[*Workload] {
	stats := make([]*stats, b.numProcs)
	for i := range b.numProcs {
		stats[i] = newStats()
	}
	w := &Workload{
		cacheSize: 50_000,
		keyspace:  500_000,
		gen:       b.gen,
		stats:     stats,
	}
	return func(yield func(*Workload) bool) {
		yield(w)
	}
}

type options struct {
	parallelism int
}

// Option is a functional option for flexible and extensible configuration of
// [*Benchmark], allowing modification of internal state or behavior during
// construction.
type Option func(*options) error

func WithParallelism(parallelism int) Option {
	return func(o *options) error {
		o.parallelism = parallelism
		return nil
	}
}

// WithOptions permits aggregating multiple options together, and is useful to
// avoid having to append options when creating helper functions or wrappers.
func WithOptions(opts ...Option) Option {
	return func(o *options) error {
		for _, opt := range opts {
			if err := opt(o); err != nil {
				return err
			}
		}
		return nil
	}
}
