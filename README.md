# DataLoader
[![GoDoc](https://godoc.org/gopkg.in/graph-gophers/dataloader.v7?status.svg)](https://pkg.go.dev/github.com/graph-gophers/dataloader/v7)
[![Build Status](https://travis-ci.org/graph-gophers/dataloader.svg?branch=master)](https://travis-ci.org/graph-gophers/dataloader)

This is an implementation of [Facebook's DataLoader](https://github.com/facebook/dataloader) in Golang.

## Install
`go get -u github.com/graph-gophers/dataloader/v7`

## Usage
```go
// setup batch function - the first Context passed to the Loader's Load
// function will be provided when the batch function is called.
// this function is registered with the Loader, and the key and value are fixed using generics.
batchFn := func(ctx context.Context, keys []int) []*dataloader.Result[*User] {
  var results []*dataloader.Result[*User]
  // do some async work to get data for specified keys
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
 *
 * loader.Load() may be called multiple times for a given batch window.
 * The first context passed to Load is the object that will be passed
 * to the batch function.
 */
thunk := loader.Load(context.TODO(), 5)
result, err := thunk()
if err != nil {
  // handle data error
}

log.Printf("value: %#v", result)
```

### Don't need/want to use context?
You're welcome to install the v1 version of this library.

## Cache
This implementation contains a very basic cache that is intended only to be used for short lived DataLoaders (i.e. DataLoaders that only exist for the life of an http request). You may use your own implementation if you want.

> it also has a `NoCache` type that implements the cache interface but all methods are noop. If you do not wish to cache anything.

## Examples
There are a few basic examples in the example folder.

## See also
- [TRACE](TRACE.md)
- [MIGRATE](MIGRATE.md)
