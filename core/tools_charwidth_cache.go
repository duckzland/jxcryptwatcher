package core

import (
	"sync"
)

var charWidthCache = &cwCache{
	store: make(map[int]float32),
}

type cwCache struct {
	mu    sync.RWMutex
	store map[int]float32
}

func (c *cwCache) Add(key int, value float32) {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.store[key]; !exists {
		c.store[key] = value
	}
}

func (c *cwCache) Get(key int) (float32, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	val, ok := c.store[key]
	return val, ok
}

func (c *cwCache) Keys() []int {
	c.mu.RLock()
	defer c.mu.RUnlock()

	keys := make([]int, 0, len(c.store))
	for k := range c.store {
		keys = append(keys, k)
	}
	return keys
}

func (c *cwCache) Has(key int) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()

	_, exists := c.store[key]
	return exists
}

func UseCharWidthCache() *cwCache {
	return charWidthCache
}
