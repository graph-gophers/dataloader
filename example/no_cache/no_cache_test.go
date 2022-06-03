package no_cache_test

import (
	"context"
	"fmt"

	dataloader "github.com/graph-gophers/dataloader/v7"
)

func ExampleNoCache() {
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

	// go-cache will automatically cleanup expired items on given duration
	cache := &dataloader.NoCache[int, *User]{}
	loader := dataloader.NewBatchedLoader(batchFunc, dataloader.WithCache[int, *User](cache))

	result, err := loader.Load(context.Background(), 5)()
	if err != nil {
		// handle error
	}

	fmt.Printf("result: %+v", result)
	// Output: result: &{ID:5 Email:john@example.com FirstName:John LastName:Smith}
}
