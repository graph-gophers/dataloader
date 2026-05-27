package dataloader_test

import (
	"context"
	"fmt"

	dataloader "github.com/graph-gophers/dataloader/v7"

	lru "github.com/hashicorp/golang-lru"
)

type lruCacheAdapter[K comparable, V any] struct {
	*lru.ARCCache
}

func (c *lruCacheAdapter[K, V]) Get(_ context.Context, key K) (dataloader.Thunk[V], bool) {
	v, ok := c.ARCCache.Get(key)
	if ok {
		return v.(dataloader.Thunk[V]), ok
	}
	return nil, ok
}

func (c *lruCacheAdapter[K, V]) Set(_ context.Context, key K, value dataloader.Thunk[V]) {
	c.ARCCache.Add(key, value)
}

func (c *lruCacheAdapter[K, V]) Delete(_ context.Context, key K) bool {
	if c.ARCCache.Contains(key) {
		c.ARCCache.Remove(key)
		return true
	}
	return false
}

func (c *lruCacheAdapter[K, V]) Clear() {
	c.ARCCache.Purge()
}

func ExampleNewBatchedLoader_golangLRU() {
	type User struct {
		ID        int
		Email     string
		FirstName string
		LastName  string
	}

	m := map[int]*User{
		5: {ID: 5, FirstName: "John", LastName: "Smith", Email: "john@example.com"},
	}

	batchFunc := func(_ context.Context, keys []int) []*dataloader.Result[*User] {
		var results []*dataloader.Result[*User]
		for _, k := range keys {
			results = append(results, &dataloader.Result[*User]{Data: m[k]})
		}
		return results
	}

	c, _ := lru.NewARC(100)
	adapter := &lruCacheAdapter[int, *User]{ARCCache: c}
	loader := dataloader.NewBatchedLoader(batchFunc, dataloader.WithCache[int, *User](adapter))

	result, err := loader.Load(context.TODO(), 5)()
	if err != nil {
		// handle error
	}

	fmt.Printf("result: %+v", result)
	// Output: result: &{ID:5 Email:john@example.com FirstName:John LastName:Smith}
}
