package dataloader

import "sync"

// The Cache interface. If a custom cache is provided, it must implement this interface.
type Cache interface {
	Get(string) (Thunk, bool)
	Set(string, Thunk)
	Delete(string) bool
	Clear()
}

// InMemoryCache is an in memory implementation of Cache interface.
// This implementation is well suited for a "per-request" dataloader
// (i.e. one that only lives for the life of an HTTP request)
// but it is not well suited for long lived cached items.
type InMemoryCache struct {
	mu    sync.RWMutex
	items map[string]Thunk
}

// NewCache constructs a new InMemoryCache.
func NewCache() *InMemoryCache {
	items := make(map[string]Thunk)
	return &InMemoryCache{
		items: items,
	}
}

// Set the `value` at `key` in the cache.
func (c *InMemoryCache) Set(key string, value Thunk) {
	c.mu.Lock()
	c.items[key] = value
	c.mu.Unlock()
}

// Get the value at `key` if it exists.
// Returns value (or nil) and whether the value was found.
func (c *InMemoryCache) Get(key string) (Thunk, bool) {
	c.mu.RLock()
	item, found := c.items[key]
	c.mu.RUnlock()

	return item, found
}

// Delete the value at `key` from the cache.
func (c *InMemoryCache) Delete(key string) bool {
	_, found := c.Get(key)

	if found {
		c.mu.Lock()
		delete(c.items, key)
		c.mu.Unlock()
	}

	return found
}

// Clear the entire cache.
func (c *InMemoryCache) Clear() {
	c.mu.Lock()
	for k := range c.items {
		delete(c.items, k)
	}
	c.mu.Unlock()
}

// NoCache implements Cache interface where all methods are noops.
// This is useful for when you don't want to cache items but still
// want to use a data loader
type NoCache struct{}

// Get is a NOOP
func (c *NoCache) Get(string) (Thunk, bool) { return nil, false }

// Set is a NOOP
func (c *NoCache) Set(string, Thunk) { return }

// Delete is a NOOP
func (c *NoCache) Delete(string) bool { return false }

// Clear is a NOOP
func (c *NoCache) Clear() { return }
