package dataloader_test

import (
	"context"
	"fmt"
	"time"

	dataloader "github.com/graph-gophers/dataloader/v7"

	cache "github.com/patrickmn/go-cache"
)

type goCacheAdapter[K comparable, V any] struct {
	c *cache.Cache
}

func (c *goCacheAdapter[K, V]) Get(_ context.Context, key K) (dataloader.Thunk[V], bool) {
	k := fmt.Sprintf("%v", key)
	v, ok := c.c.Get(k)
	if ok {
		return v.(dataloader.Thunk[V]), ok
	}
	return nil, ok
}

func (c *goCacheAdapter[K, V]) Set(_ context.Context, key K, value dataloader.Thunk[V]) {
	k := fmt.Sprintf("%v", key)
	c.c.Set(k, value, 0)
}

func (c *goCacheAdapter[K, V]) Delete(_ context.Context, key K) bool {
	k := fmt.Sprintf("%v", key)
	if _, found := c.c.Get(k); found {
		c.c.Delete(k)
		return true
	}
	return false
}

func (c *goCacheAdapter[K, V]) Clear() {
	c.c.Flush()
}

func ExampleNewBatchedLoader_goCache() {
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

	c := cache.New(15*time.Minute, 15*time.Minute)
	adapter := &goCacheAdapter[int, *User]{c}
	loader := dataloader.NewBatchedLoader(batchFunc, dataloader.WithCache[int, *User](adapter))

	result, err := loader.Load(context.Background(), 5)()
	if err != nil {
		// handle error
	}

	fmt.Printf("result: %+v", result)
	// Output: result: &{ID:5 Email:john@example.com FirstName:John LastName:Smith}
}
