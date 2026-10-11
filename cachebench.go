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
	gen       *generator.FiniteZipf
	workloads []*Workload
}

func New(b *testing.B, opts ...Option) *Benchmark {
	cfg := options{
		parallelism: 1,
	}
	if err := WithOptions(opts...)(&cfg); err != nil {
		b.Fatal(err)
	}

	numProcs := runtime.GOMAXPROCS(0) * cfg.parallelism

	cacheSizes := []uint64{
		50_000,
	}

	var workloads []*Workload
	for _, cacheSize := range cacheSizes {
		keyspace := 10 * cacheSize
		gen, err := generator.NewFiniteZipf(keyspace, 0.75, generator.WithSeed(1))
		if err != nil {
			b.Fatal(err)
		}
		stats := make([]*stats, numProcs)
		for i := range numProcs {
			stats[i] = newStats()
		}
		workloads = append(workloads, &Workload{
			cacheSize: int(cacheSize),
			keyspace:  keyspace,
			gen:       gen,
			stats:     stats,
		})
	}

	b.ReportAllocs()
	b.SetParallelism(cfg.parallelism)
	return &Benchmark{
		workloads: workloads,
	}
}

func (b *Benchmark) Workloads() iter.Seq[*Workload] {
	return func(yield func(*Workload) bool) {
		for _, workload := range b.workloads {
			if !yield(workload) {
				return
			}
		}
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
