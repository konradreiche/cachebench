// Package generator generates random keys for cachebench benchmarks.
package generator

import (
	"math"
	"math/rand/v2"
	"sort"
	"strconv"
	"sync"
)

type FiniteZipf struct {
	cdf []float64

	mu  sync.Mutex
	rng *rand.Rand
}

func NewFiniteZipf(keyspace uint64, α float64, opts ...Option) (*FiniteZipf, error) {
	cfg := options{
		seed1: rand.Uint64(),
		seed2: rand.Uint64(),
	}
	if err := WithOptions(opts...)(&cfg); err != nil {
		return nil, err
	}

	cdf := make([]float64, keyspace)
	var total float64
	for i := range keyspace {
		total += math.Pow(float64(i+1), -α)
		cdf[i] = total
	}

	for i := range cdf {
		cdf[i] /= total
	}

	return &FiniteZipf{
		rng: rand.New(rand.NewPCG(cfg.seed1, cfg.seed2)),
		cdf: cdf,
	}, nil
}

func (z *FiniteZipf) Next() string {
	z.mu.Lock()
	u := z.rng.Float64()
	z.mu.Unlock()
	i := sort.SearchFloat64s(z.cdf, u)
	return strconv.FormatUint(uint64(i+1), 36)
}

type options struct {
	seed1 uint64
	seed2 uint64
}

// Option is a functional option for flexible and extensible configuration of
// [*FiniteZipf], allowing modification of internal state or behavior during
// construction.
type Option func(*options) error

func WithSeed(seed uint64) Option {
	return func(o *options) error {
		o.seed1 = seed
		o.seed2 = seed
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
