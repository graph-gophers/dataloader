// package ttl_cache_test contains an exmaple of using go-cache as a long term cache solution for dataloader.
package ttl_cache_test

import (
	"context"
	"fmt"
	"time"

	dataloader "github.com/graph-gophers/dataloader/v7"

	cache "github.com/patrickmn/go-cache"
)

// Cache implements the dataloader.Cache interface
type Cache[K comparable, V any] struct {
	c *cache.Cache
}

// Get gets a value from the cache
func (c *Cache[K, V]) Get(_ context.Context, key K) (dataloader.Thunk[V], bool) {
	k := fmt.Sprintf("%v", key) // convert the key to string because the underlying library doesn't support Generics yet
	v, ok := c.c.Get(k)
	if ok {
		return v.(dataloader.Thunk[V]), ok
	}
	return nil, ok
}

// Set sets a value in the cache
func (c *Cache[K, V]) Set(_ context.Context, key K, value dataloader.Thunk[V]) {
	k := fmt.Sprintf("%v", key) // convert the key to string because the underlying library doesn't support Generics yet
	c.c.Set(k, value, 0)
}

// Delete deletes and item in the cache
func (c *Cache[K, V]) Delete(_ context.Context, key K) bool {
	k := fmt.Sprintf("%v", key) // convert the key to string because the underlying library doesn't support Generics yet
	if _, found := c.c.Get(k); found {
		c.c.Delete(k)
		return true
	}
	return false
}

// Clear clears the cache
func (c *Cache[K, V]) Clear() {
	c.c.Flush()
}

func ExampleTTLCache() {
	type User struct {
		ID        int
		Email     string
		FirstName string
		LastName  string
	}

	m := map[int]*User{
		5: &User{ID: 5, FirstName: "John", LastName: "Smith", Email: "john@example.com"},
	}

	batchFunc := func(_ context.Context, keys []int) []*dataloader.Result[*User] {
		var results []*dataloader.Result[*User]
		// do some pretend work to resolve keys
		for _, k := range keys {
			results = append(results, &dataloader.Result[*User]{Data: m[k]})
		}
		return results
	}

	// go-cache will automaticlly cleanup expired items on given diration
	c := cache.New(15*time.Minute, 15*time.Minute)
	cache := &Cache[int, *User]{c}
	loader := dataloader.NewBatchedLoader(batchFunc, dataloader.WithCache[int, *User](cache))

	// immediately call the future function from loader
	result, err := loader.Load(context.Background(), 5)()
	if err != nil {
		// handle error
	}

	fmt.Printf("result: %+v", result)
	// Output: result: &{ID:5 Email:john@example.com FirstName:John LastName:Smith}
}
