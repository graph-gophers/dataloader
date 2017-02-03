// This is an exmaple of using go-cache as a long term cache solution for
// dataloader.
package main

import (
	"fmt"

	lru "github.com/hashicorp/golang-lru"
	"github.com/nicksrandall/dataloader"
)

type Cache struct {
	*lru.ARCCache
}

func (c *Cache) Get(key string) (dataloader.Thunk, bool) {
	v, ok := c.ARCCache.Get(key)
	if ok {
		return v.(dataloader.Thunk), ok
	}
	return nil, ok
}

func (c *Cache) Set(key string, value dataloader.Thunk) {
	c.ARCCache.Add(key, value)
}

func (c *Cache) Delete(key string) bool {
	if c.ARCCache.Contains(key) {
		c.ARCCache.Remove(key)
		return true
	}
	return false
}

func (c *Cache) Clear() {
	c.ARCCache.Purge()
}

func main() {
	// go-cache will automaticlly cleanup expired items on given diration
	c, _ := lru.NewARC(100)
	cache := &Cache{c}
	loader := dataloader.NewBatchedLoader(batchFunc, dataloader.WithCache(cache))

	// immediately call the future function from loader
	result, err := loader.Load("some key")()
	if err != nil {
		// handle error
	}

	fmt.Printf("identity: %s\n", result)
}

func batchFunc(keys []string) []*dataloader.Result {
	var results []*dataloader.Result
	// do some pretend work to resolve keys
	for _, key := range keys {
		results = append(results, &dataloader.Result{key, nil})
	}
	return results
}
