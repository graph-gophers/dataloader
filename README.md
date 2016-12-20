# DataLoader
[![GoDoc](https://godoc.org/github.com/nicksrandall/dataloader?status.svg)](https://godoc.org/github.com/nicksrandall/dataloader)
[![Build Status](https://travis-ci.org/nicksrandall/dataloader.svg?branch=master)](https://travis-ci.org/nicksrandall/dataloader)
[![Coverage](http://gocover.io/_badge/github.com/nicksrandall/dataloader)](http://gocover.io/github.com/nicksrandall/dataloader)

This is an implementation of [Facebook's DataLoader](https://github.com/facebook/dataloader) in Golang.

## Status
This project is a work in progress. Feedback is encouraged.

## Usage
```go
// setup batch function
batchFn := func(keys []string) []dataloader.Result {
  var results []dataloader.Result
  // do some aync work to get data for specified keys
  // append to this list resolved values
  return results
}

// create Loader with an in-memory cache
loader := dataloader.NewBatchedLoader(batchFn)

/**
 * Use loader
 *
 * A thunk is a function returned from a function that is a 
 * closure over a value (in this case an interface value and error).
 * When called, it will block until the value is resolved.
 */
thunk := loader.Load("key1")
result, err := thunk()
if err != nil {
  // handle data error
}

log.Printf("value: %#v", result)
```

## Cache
This implementation contains a very basic cache that is intended only to be used for short lived DataLoaders (i.e. DataLoaders that ony exsist for the life of an http request). You may use your own implementation if you want.

> it also has a `NoCache` type that implements the cache interface but all methods are noop. If you do not wish to cache anyting.

## Examples
There are a few basic examples in the example folder.
