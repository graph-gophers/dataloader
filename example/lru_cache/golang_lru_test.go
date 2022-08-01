// package lru_cache_test contains an example of using go-cache as a long term cache solution for dataloader.
package lru_cache_test

import (
	"context"
	"fmt"

	lru "github.com/hashicorp/golang-lru"

	dataloader "github.com/graph-gophers/dataloader/v8"
)

// Cache implements the dataloader.Cache interface
type cache[K any, V any] struct {
	*lru.ARCCache
}

// Get gets an item from the cache
func (c *cache[K, V]) Get(_ context.Context, key dataloader.Key[K]) (dataloader.Thunk[V], bool) {
	v, ok := c.ARCCache.Get(key.String())
	if ok {
		return v.(dataloader.Thunk[V]), ok
	}
	return nil, ok
}

// Set sets an item in the cache
func (c *cache[K, V]) Set(_ context.Context, key dataloader.Key[K], value dataloader.Thunk[V]) {
	c.ARCCache.Add(key.String(), value)
}

// Delete deletes an item in the cache
func (c *cache[K, V]) Delete(_ context.Context, key dataloader.Key[K]) bool {
	if c.ARCCache.Contains(key.String()) {
		c.ARCCache.Remove(key.String())
		return true
	}
	return false
}

// Clear clears the cache
func (c *cache[K, V]) Clear() {
	c.ARCCache.Purge()
}

func ExampleGolangLRU() {
	type User struct {
		ID        int
		Email     string
		FirstName string
		LastName  string
	}

	m := map[int]*User{
		5: {ID: 5, FirstName: "John", LastName: "Smith", Email: "john@example.com"},
	}

	batchFunc := func(_ context.Context, keys dataloader.Keys[int]) []*dataloader.Result[*User] {
		var results []*dataloader.Result[*User]
		// do some pretend work to resolve keys
		for _, k := range keys {
			results = append(results, &dataloader.Result[*User]{Data: m[k.Raw()]})
		}
		return results
	}

	// go-cache will automaticlly cleanup expired items on given duration.
	c, _ := lru.NewARC(100)
	cache := &cache[int, *User]{ARCCache: c}
	loader := dataloader.NewBatchedLoader(batchFunc, dataloader.WithCache[int, *User](cache))

	// immediately call the future function from loader
	result, err := loader.Load(context.TODO(), dataloader.KeyOf(5))()
	if err != nil {
		// handle error
	}

	fmt.Printf("result: %+v", result)
	// Output: result: &{ID:5 Email:john@example.com FirstName:John LastName:Smith}
}
