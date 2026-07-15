// Package lfu provides a cache with LFU replacement policy.
package lfu

import (
	"github.com/konradreiche/cachebench/benchmark/internal/lfu/internal/lfuheap"
	"github.com/konradreiche/cachebench/benchmark/internal/lfu/internal/lfupid"
)

type Cache[K comparable, V any] struct {
	size int

	lfuHeap *lfuheap.Cache[K, V]
	lfuPID  *lfupid.Cache[K, V]
}

func New[K comparable, V any](size int, opts ...Option) (*Cache[K, V], error) {
	cfg := options{}
	if err := WithOptions(opts...)(&cfg); err != nil {
		return nil, err
	}
	lfuHeap, err := lfuheap.New[K, V](size)
	if err != nil {
		return nil, err
	}
	cache := &Cache[K, V]{
		size:    size,
		lfuHeap: lfuHeap,
	}
	if cfg.usePID {
		lfuPID, err := lfupid.New[K, V](size)
		if err != nil {
			return nil, err
		}
		cache.lfuPID = lfuPID
	}
	return cache, nil
}

func (c *Cache[K, V]) Store(key K, value V) {
	if c.lfuPID != nil {
		c.lfuPID.Store(key, value)
		return
	}
	c.lfuHeap.Store(key, value)
}

func (c *Cache[K, V]) Load(key K) (V, bool) {
	if c.lfuPID != nil {
		return c.lfuPID.Load(key)
	}
	return c.lfuHeap.Load(key)
}

type options struct {
	usePID bool
}

// Option is a functional option for flexible and extensible configuration of
// [*Cache], allowing modification of internal state or behavior during
// construction.
type Option func(*options) error

func WithUsePID(enabled bool) Option {
	return func(o *options) error {
		o.usePID = enabled
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
