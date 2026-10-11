// Package lfupid provides a cache with LFU replacement policy.
package lfupid

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/konradreiche/pid"
	"github.com/puzpuzpuz/xsync/v4"
)

type Cache[K comparable, V any] struct {
	size               int
	data               *xsync.Map[K, *item[K, V]]
	analyzer           *Analyzer
	backgroundEviction bool

	mu         sync.Mutex
	controller *pid.Controller
	scoreMin   float64
	updatedAt  time.Time
}

func New[K comparable, V any](size int, opts ...Option) (*Cache[K, V], error) {
	cfg := options{
		backgroundEviction: false,
	}
	if err := WithOptions(opts...)(&cfg); err != nil {
		return nil, err
	}
	controller, err := pid.New(
		pid.WithProportionalGain(16.0),
	)
	if err != nil {
		return nil, err
	}

	analyzer, err := NewAnalyzer("lfu")
	if err != nil {
		return nil, err
	}
	cache := &Cache[K, V]{
		size:               size,
		data:               xsync.NewMap[K, *item[K, V]](xsync.WithPresize(size)),
		analyzer:           analyzer,
		backgroundEviction: cfg.backgroundEviction,
		updatedAt:          time.Now(),
		controller:         controller,
	}
	if cfg.backgroundEviction {
		go cache.evictTask()
	}
	return cache, nil
}

func (c *Cache[K, V]) Store(key K, value V) {
	var frequency atomic.Uint64
	frequency.Add(1)
	entry := &item[K, V]{
		frequency: &frequency,
		value:     value,
	}
	c.data.Store(key, entry)
	if c.backgroundEviction {
		return
	}
	if c.data.Size() > c.size {
		c.evict()
	}
}

func (c *Cache[K, V]) evictTask() {
	for {
		time.Sleep(100 * time.Millisecond)
		c.evict()
	}
}

func (c *Cache[K, V]) evict() {
	if c.data.Size() <= c.size {
		return
	}
	scoreMin := c.scoreMin
	c.data.RangeRelaxed(func(key K, value *item[K, V]) bool {
		if float64(value.frequency.Load()) < scoreMin {
			c.data.Delete(key)
		}
		return true
	})

	current := float64(c.data.Size()) / float64(c.size)
	signal := c.controller.Update(0.95, current, time.Since(c.updatedAt))
	c.scoreMin -= signal
	c.updatedAt = time.Now()
	c.analyzer.log(c.updatedAt, current, signal, c.scoreMin)
}

func (c *Cache[K, V]) Load(key K) (V, bool) {
	item, ok := c.data.Load(key)
	if !ok {
		return *new(V), false
	}
	item.frequency.Add(1)
	return item.value, true
}

func (c *Cache[K, V]) snapshot() {
	frequencyBuckets := make(map[uint64]int)
	c.data.Range(func(key K, value *item[K, V]) bool {
		freq := value.frequency.Load()
		frequencyBuckets[freq]++
		return true
	})
}

type item[K comparable, V any] struct {
	value     V
	frequency *atomic.Uint64
	index     int
}

type options struct {
	backgroundEviction bool
}

// Option is a functional option for flexible and extensible configuration of
// [Cache], allowing modification of internal state or behavior during
// construction.
type Option func(*options) error

func WithBackgroundEviction(enabled bool) Option {
	return func(o *options) error {
		o.backgroundEviction = enabled
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
