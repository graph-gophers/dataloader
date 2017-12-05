// This is an exmaple of using go-cache as a long term cache solution for
// dataloader.
package main

import (
	"context"
	"fmt"
	"time"

	"github.com/nicksrandall/dataloader"
	cache "github.com/patrickmn/go-cache"
)

// Cache implements the dataloader.Cache interface
type Cache struct {
	c *cache.Cache
}

// Get gets a value from the cache
func (c *Cache) Get(_ context.Context, key interface{}) (dataloader.Thunk, bool) {
	v, ok := c.c.Get(key.(string))
	if ok {
		return v.(dataloader.Thunk), ok
	}
	return nil, ok
}

// Set sets a value in the cache
func (c *Cache) Set(_ context.Context, key interface{}, value dataloader.Thunk) {
	c.c.Set(key.(string), value, 0)
}

// Delete deletes and item in the cache
func (c *Cache) Delete(_ context.Context, key interface{}) bool {
	if _, found := c.c.Get(key.(string)); found {
		c.c.Delete(key.(string))
		return true
	}
	return false
}

// Clear clears the cache
func (c *Cache) Clear() {
	c.c.Flush()
}

func main() {
	// go-cache will automaticlly cleanup expired items on given diration
	c := cache.New(15*time.Minute, 15*time.Minute)
	cache := &Cache{c}
	loader := dataloader.NewBatchedLoader(batchFunc, dataloader.WithCache(cache))

	// immediately call the future function from loader
	result, err := loader.Load(context.TODO(), "some key")()
	if err != nil {
		// handle error
	}

	fmt.Printf("identity: %s\n", result)
}

func batchFunc(_ context.Context, keys []interface{}) []*dataloader.Result {
	var results []*dataloader.Result
	// do some pretend work to resolve keys
	for _, key := range keys {
		results = append(results, &dataloader.Result{key, nil})
	}
	return results
}
