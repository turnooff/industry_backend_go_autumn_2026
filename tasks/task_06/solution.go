package main

import "container/list"

type entry[K comparable, V any] struct {
	key   K
	value V
}
type LRUCache[K comparable, V any] struct {
	capacity int
	ll       list.List
	items    map[K]*list.Element
}

func NewLRUCache[K comparable, V any](capacity int) *LRUCache[K, V] {
	lru := &LRUCache[K, V]{}

	lru.capacity = capacity
	lru.ll = list.List{}
	lru.items = make(map[K]*list.Element)

	return lru
}
func (c *LRUCache[K, V]) Get(key K) (value V, ok bool) {
	if c.capacity <= 0 {
		return value, false
	}

	v, ok := c.items[key]
	if ok {
		c.ll.MoveToFront(v)
		value = v.Value.(entry[K, V]).value
		return value, ok
	}

	return value, false
}
func (c *LRUCache[K, V]) Set(key K, value V) {
	if c.capacity <= 0 {
		return
	}

	e, ok := c.items[key]
	if ok {
		e.Value = entry[K, V]{key, value}
		return
	}

	if c.ll.Len() == c.capacity {
		lastE := c.ll.Back()
		lastK := lastE.Value.(entry[K, V]).key
		c.ll.Remove(lastE)
		delete(c.items, lastK)
	}
	newE := c.ll.PushFront(entry[K, V]{key, value})
	c.items[key] = newE
}

type LRU[K comparable, V any] interface {
	Get(K) (V, bool)
	Set(K, V)
}
