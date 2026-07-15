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
	size int
	data *xsync.Map[K, *item[K, V]]

	mu         sync.Mutex
	controller *pid.Controller
	scoreMin   float64
	updatedAt  time.Time
}

func New[K comparable, V any](size int) (*Cache[K, V], error) {
	controller, err := pid.New(
		pid.WithZieglerNicholsMethod(1.0, 1.0),
	)
	if err != nil {
		return nil, err
	}

	return &Cache[K, V]{
		size:       size,
		data:       xsync.NewMap[K, *item[K, V]](xsync.WithPresize(size)),
		updatedAt:  time.Now(),
		controller: controller,
	}, nil
}

func (c *Cache[K, V]) Store(key K, value V) {
	var frequency atomic.Uint64
	frequency.Add(1)
	entry := &item[K, V]{
		key:       key,
		frequency: &frequency,
		value:     value,
	}
	c.data.Store(key, entry)
	if c.data.Size() > c.size {
		c.mu.Lock()
		c.evict()
		c.mu.Unlock()
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
}

func (c *Cache[K, V]) Load(key K) (V, bool) {
	item, ok := c.data.Load(key)
	if !ok {
		return *new(V), false
	}
	item.frequency.Add(1)
	return item.value, true
}

type item[K comparable, V any] struct {
	key       K
	value     V
	frequency *atomic.Uint64
	index     int
}
