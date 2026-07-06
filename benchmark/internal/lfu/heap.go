package lfu

import "container/heap"

type item[K comparable, V any] struct {
	key       K
	value     V
	frequency int
	lastUsed  int
	index     int
}

type MinHeap[K comparable, V any] []*item[K, V]

func (h MinHeap[K, V]) Len() int { return len(h) }

func (h MinHeap[K, V]) Less(i, j int) bool {
	if h[i].frequency != h[j].frequency {
		return h[i].frequency < h[j].frequency
	}
	return h[i].lastUsed < h[j].lastUsed
}

func (h MinHeap[K, V]) Swap(i, j int) {
	h[i], h[j] = h[j], h[i]
	h[i].index = i
	h[j].index = j
}

func (h *MinHeap[K, V]) Push(x any) {
	n := len(*h)
	item := x.(*item[K, V])
	item.index = n
	*h = append(*h, item)
}

func (h *MinHeap[K, V]) Pop() any {
	old := *h
	n := len(old)
	item := old[n-1]
	old[n-1] = nil
	item.index = -1
	*h = old[0 : n-1]
	return item
}

func (h *MinHeap[K, V]) update(item *item[K, V], value V, priority, lastUsed int) {
	item.value = value
	item.frequency = priority
	item.lastUsed = lastUsed
	heap.Fix(h, item.index)
}
