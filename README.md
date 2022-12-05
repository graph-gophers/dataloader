# DataLoader
[![GoDoc](https://godoc.org/gopkg.in/graph-gophers/dataloader.v3?status.svg)](https://godoc.org/github.com/graph-gophers/dataloader)
[![Build Status](https://travis-ci.org/graph-gophers/dataloader.svg?branch=master)](https://travis-ci.org/graph-gophers/dataloader)

This is an implementation of [Facebook's DataLoader](https://github.com/facebook/dataloader) in Golang.

## Install
`go get -u github.com/graph-gophers/dataloader/v7

## Usage
```go
// setup batch function - the first Context passed to the Loader's Load
// function will be provided when the batch function is called.
batchFn := func(ctx context.Context, keys []string) []*dataloader.Result[*model.User] {
  var results []*dataloader.Result[(model.User]
  // do some async work to get data for specified keys
  // append to this list resolved values
  // select * form user where id in (...) -- keys
  // scan rows make it to results
  return results
}

// create Loader with an in-memory cache
var userLoader *dataloader.Loader[string, *model.User)
userLoader := dataloader.NewBatchedLoader(batchFn)

// create Loader with no cache
userLoader := dataloader.NewBatchedLoader(batchFn, dataloader.WithNoCache[string, *model.User]())

/**
 * Use loader
 *
 * A thunk is a function returned from a function that is a
 * closure over a value (in this case an interface value and error).
 * When called, it will block until the value is resolved.
 *
 * userLoader.Load() may be called multiple times for a given batch window.
 * The first context passed to Load is the object that will be passed
 * to the batch function.
 */
thunk := userLoader.Load(context.Background(), "key1")
results, err := thunk()
if err != nil {
  // handle data error
}

log.Printf("value: %#v", results)
```

### Don't need/want to use context?
You're welcome to install the v1 version of this library.

## Cache
This implementation contains a very basic cache that is intended only to be used for short lived DataLoaders (i.e. DataLoaders that only exist for the life of an http request). You may use your own implementation if you want.

> it also has a `NoCache` type that implements the cache interface but all methods are noop. If you do not wish to cache anything.

## Examples
There are a few basic examples in the example folder.
