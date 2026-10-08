package main

type Cache[K comparable, V any] struct {
	capacity int
	items    map[K]V
}

func NewCache[K comparable, V any](capacity int) *Cache[K, V] {
	cache := new(Cache[K, V])
	cache.capacity = capacity
	cache.items = make(map[K]V)

	return cache
}
func (c *Cache[K, V]) Get(k K) (v V, ok bool) {
	if c.capacity <= 0 {
		return v, false
	}

	v, ok = c.items[k]
	return v, ok
}
func (c *Cache[K, V]) Set(k K, v V) bool {
	if c.capacity <= 0 {
		return false
	}

	_, ok := c.items[k]

	if !ok && len(c.items) == c.capacity {
		return false
	}

	c.items[k] = v
	return true
}
