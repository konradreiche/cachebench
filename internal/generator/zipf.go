// Package generator generates random keys for cachebench benchmarks.
package generator

import (
	"iter"
	"math"
	"math/rand/v2"
	"sort"
	"strconv"
)

type FiniteZipf struct {
	rng *rand.Rand
	cdf []float64
}

func NewFiniteZipf(keyspace uint64, α float64) *FiniteZipf {
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
		rng: rand.New(rand.NewPCG(rand.Uint64(), rand.Uint64())),
		cdf: cdf,
	}
}

func (z *FiniteZipf) Next() iter.Seq[string] {
	return func(yield func(string) bool) {
		u := z.rng.Float64()
		i := sort.SearchFloat64s(z.cdf, u)
		if !yield(strconv.FormatUint(uint64(i+1), 36)) {
			return
		}
	}
}
