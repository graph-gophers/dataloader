package dataloader

import "sync"

// If a custom cache is provided, it must implement this interface.
type Cache interface {
	Get(string) (interface{}, bool)
	Set(string, interface{})
	Delete(string)
	Clear()
}

// In memory implementation of Cache interace.
// this simple implementation is well suited for
// a "per-request" dataloader (i.e. one that only lives
// for the life of an http request) but it not well suited
// for long lived cached items.
type InMemoryCache struct {
	items map[string]interface{}
	mu    sync.RWMutex
}

// NewCache constructs a new InMemoryCache
func NewCache() *InMemoryCache {
	items := make(map[string]interface{})
	return &InMemoryCache{
		items: items,
	}
}

// sets the `value` at `key` in the cache
func (c *InMemoryCache) Set(key string, value interface{}) {
	c.mu.Lock()
	c.items[key] = value
	c.mu.Unlock()
}

// gets the value at `key` if it exsits, returns value (or nil) and bool
// indicating of value was found
func (c *InMemoryCache) Get(key string) (interface{}, bool) {
	c.mu.RLock()

	item, found := c.items[key]
	if !found {
		c.mu.RUnlock()
		return nil, false
	}

	c.mu.RUnlock()
	return item, true
}

// deletes item at `key` from cache
func (c *InMemoryCache) Delete(key string) {
	c.mu.Lock()
	delete(c.items, key)
	c.mu.Unlock()
}

// clears the entire cache
func (c *InMemoryCache) Clear() {
	c.mu.Lock()
	c.items = map[string]interface{}{}
	c.mu.Unlock()
}
