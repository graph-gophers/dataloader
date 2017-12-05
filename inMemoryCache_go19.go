// +build go1.9

package dataloader

import (
	"context"
	"sync"
)

// InMemoryCache is an in memory implementation of Cache interface.
// this simple implementation is well suited for
// a "per-request" dataloader (i.e. one that only lives
// for the life of an http request) but it not well suited
// for long lived cached items.
type InMemoryCache struct {
	items *sync.Map
}

// NewCache constructs a new InMemoryCache
func NewCache() *InMemoryCache {
	return &InMemoryCache{
		items: &sync.Map{},
	}
}

// Set sets the `value` at `key` in the cache
func (c *InMemoryCache) Set(_ context.Context, key interface{}, value Thunk) {
	c.items.Store(key, value)
}

// Get gets the value at `key` if it exsits, returns value (or nil) and bool
// indicating of value was found
func (c *InMemoryCache) Get(_ context.Context, key interface{}) (Thunk, bool) {
	item, found := c.items.Load(key)
	if !found {
		return nil, false
	}

	return item.(Thunk), true
}

// Delete deletes item at `key` from cache
func (c *InMemoryCache) Delete(_ context.Context, key interface{}) bool {
	if _, found := c.items.Load(key); found {
		c.items.Delete(key)
		return true
	}
	return false
}

// Clear clears the entire cache
func (c *InMemoryCache) Clear() {
	c.items.Range(func(key, _ interface{}) bool {
		c.items.Delete(key)
		return true
	})
}
