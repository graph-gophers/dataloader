package main

import (
	"fmt"
	"time"

	"github.com/nicksrandall/dataloader"
)

func main() {
	// go-cache will automaticlly cleanup expired items on given diration
	c := &dataloader.NoCache{}
	cache := &Cache{c}
	loader := dataloader.NewBatchedLoader(batchFunc, time.Duration(16*time.Millisecond), cache, 0)

	result := <-loader.Load("some key")
	if result.Error != nil {
		// handle error
	}

	fmt.Printf("identity: %s\n", result.Data)
}

func batchFunc(keys []string) []dataloader.Result {
	var results []dataloader.Result
	// do some pretend work to resolve keys
	for _, key := range keys {
		results = append(results, dataloader.Result{key, nil})
	}
	return results
}
