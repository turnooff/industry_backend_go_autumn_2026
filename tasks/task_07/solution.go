package main

import (
	"container/list"
	"sync"
)

type LRU[K comparable, V any] interface {
	Get(key K) (value V, ok bool)
	Set(key K, value V)
}
type entry[K comparable, V any] struct {
	key   K
	value V
}
type LRUCache[K comparable, V any] struct {
	capacity int
	mu       sync.Mutex
	ll       list.List
	items    map[K]*list.Element
}

func NewLRUCache[K comparable, V any](capacity int) *LRUCache[K, V] {
	newLRU := &LRUCache[K, V]{}

	newLRU.capacity = capacity
	newLRU.mu = sync.Mutex{}
	newLRU.ll = list.List{}
	newLRU.items = make(map[K]*list.Element)

	return newLRU
}
func (c *LRUCache[K, V]) Get(key K) (value V, ok bool) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.capacity <= 0 {
		return value, false
	}

	e, ok := c.items[key]
	if ok {
		c.ll.MoveToFront(e)
		value = e.Value.(entry[K, V]).value
		return value, true
	}

	return value, false
}
func (c *LRUCache[K, V]) Set(key K, value V) {
	c.mu.Lock()
	defer c.mu.Unlock()

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
