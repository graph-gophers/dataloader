// This is an exmaple of using go-cache as a long term cache solution for
// dataloader.
package main

import (
	"fmt"
	"time"

	"github.com/nicksrandall/dataloader"
	cache "github.com/patrickmn/go-cache"
)

type Cache struct {
	c *cache.Cache
}

func (c *Cache) Get(key string) (dataloader.Thunk, bool) {
	v, ok := c.c.Get(key)
	if ok {
		return v.(dataloader.Thunk), ok
	}
	return nil, ok
}

func (c *Cache) Set(key string, value dataloader.Thunk) {
	c.c.Set(key, value, 0)
}

func (c *Cache) Delete(key string) {
	c.c.Delete(key)
}

func (c *Cache) Clear() {
	c.c.Flush()
}

func main() {
	// go-cache will automaticlly cleanup expired items on given diration
	c := cache.New(time.Duration(15*time.Minute), time.Duration(15*time.Minute))
	cache := &Cache{c}
	loader := dataloader.NewBatchedLoader(batchFunc, dataloader.WithCache(cache))

	// immediately call the future function from loader
	result := loader.Load("some key")()
	if result.Error != nil {
		// handle error
	}

	fmt.Printf("identity: %s\n", result.Data)
}

func batchFunc(keys []string) []*dataloader.Result {
	var results []*dataloader.Result
	// do some pretend work to resolve keys
	for _, key := range keys {
		results = append(results, &dataloader.Result{key, nil})
	}
	return results
}
