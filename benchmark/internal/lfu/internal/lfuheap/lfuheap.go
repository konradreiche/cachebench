// Package lfuheap provides a cache with LFU replacement policy.
package lfuheap

import (
	"container/heap"
	"sync"
)

type Cache[K comparable, V any] struct {
	size int

	mu      sync.Mutex
	data    map[K]*item[K, V]
	clock   int
	minHeap *MinHeap[K, V]
}

func New[K comparable, V any](size int) (*Cache[K, V], error) {
	return &Cache[K, V]{
		size:    size,
		data:    make(map[K]*item[K, V]),
		minHeap: &MinHeap[K, V]{},
	}, nil
}

func (c *Cache[K, V]) Store(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if entry, ok := c.data[key]; ok {
		c.minHeap.update(entry, value, entry.frequency, entry.lastUsed)
		return
	}

	entry := &item[K, V]{
		key:       key,
		frequency: 1,
		lastUsed:  c.clock,
		value:     value,
	}
	c.data[key] = entry
	heap.Push(c.minHeap, entry)

	for len(c.data) > c.size {
		remove := heap.Pop(c.minHeap).(*item[K, V])
		delete(c.data, remove.key)
	}
}

func (c *Cache[K, V]) Load(key K) (V, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	item, ok := c.data[key]
	if !ok {
		return *new(V), false
	}
	c.clock++
	item.frequency++
	c.minHeap.update(item, item.value, item.frequency, item.lastUsed)
	return item.value, true
}
