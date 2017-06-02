package main

import (
	"context"
	"fmt"

	"github.com/nicksrandall/dataloader"
)

func main() {
	// go-cache will automaticlly cleanup expired items on given diration
	cache := &dataloader.NoCache{}
	loader := dataloader.NewBatchedLoader(batchFunc, dataloader.WithCache(cache))

	result, err := loader.Load(context.TODO(), "some key")()
	if err != nil {
		// handle error
	}

	fmt.Printf("identity: %s\n", result)
}

func batchFunc(_ context.Context, keys []string) []*dataloader.Result {
	var results []*dataloader.Result
	// do some pretend work to resolve keys
	for _, key := range keys {
		results = append(results, &dataloader.Result{key, nil})
	}
	return results
}
