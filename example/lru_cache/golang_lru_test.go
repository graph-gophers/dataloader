package lru_cache_test

// This is an exmaple of using go-cache as a long term cache solution for
// dataloader.

import (
	"context"
	"fmt"

	dataloader "github.com/graph-gophers/dataloader/v6"
	lru "github.com/hashicorp/golang-lru"
)

// Cache implements the dataloader.Cache interface
type cache struct {
	*lru.ARCCache
}

// Get gets an item from the cache
func (c *cache) Get(_ context.Context, key dataloader.Key) (dataloader.Thunk, bool) {
	v, ok := c.ARCCache.Get(key)
	if ok {
		return v.(dataloader.Thunk), ok
	}
	return nil, ok
}

// Set sets an item in the cache
func (c *cache) Set(_ context.Context, key dataloader.Key, value dataloader.Thunk) {
	c.ARCCache.Add(key, value)
}

// Delete deletes an item in the cache
func (c *cache) Delete(_ context.Context, key dataloader.Key) bool {
	if c.ARCCache.Contains(key) {
		c.ARCCache.Remove(key)
		return true
	}
	return false
}

// Clear cleasrs the cache
func (c *cache) Clear() {
	c.ARCCache.Purge()
}

func ExampleGolangLRU() {
	// go-cache will automaticlly cleanup expired items on given duration.
	c, _ := lru.NewARC(100)
	cache := &cache{ARCCache: c}
	loader := dataloader.NewBatchedLoader(batchFunc, dataloader.WithCache(cache))

	// immediately call the future function from loader
	result, err := loader.Load(context.TODO(), dataloader.StringKey("some key"))()
	if err != nil {
		// handle error
	}

	fmt.Printf("identity: %s", result)
	// Output: identity: some key
}

func batchFunc(_ context.Context, keys dataloader.Keys) []*dataloader.Result {
	var results []*dataloader.Result
	// do some pretend work to resolve keys
	for _, key := range keys {
		results = append(results, &dataloader.Result{Data: key.String()})
	}
	return results
}
