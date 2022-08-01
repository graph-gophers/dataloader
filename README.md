# DataLoader
[![GoDoc](https://godoc.org/gopkg.in/graph-gophers/dataloader.v7?status.svg)](https://godoc.org/github.com/graph-gophers/dataloader)
[![Build Status](https://travis-ci.org/graph-gophers/dataloader.svg?branch=master)](https://travis-ci.org/graph-gophers/dataloader)

This is an implementation of [Facebook's DataLoader](https://github.com/facebook/dataloader) in Golang.

## Install
`go get -u github.com/graph-gophers/dataloader/v8`

## Usage
```go
// setup batch function - the first Context passed to the Loader's Load
// function will be provided when the batch function is called.
batchFn := func(ctx context.Context, keys dataloader.Keys[string]) []*dataloader.Result[any] {
  var results []*dataloader.Result[any]
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
thunk := loader.Load(context.TODO(), dataloader.KeyOf("key1")) // KeyOf is a convenience method that wraps any comparable type to implement `Key` interface
result, err := thunk()
if err != nil {
  // handle data error
}

log.Printf("value: %#v", result)
```

### Don't need/want to use context?
You're welcome to install the `v1` version of this library.

### Don't need/want to use type parameters?
Please feel free to use `v6` version of this library.

### Don't need/want to use Key/Keys interface?
Just use the `v7` version of this library. This completely removes the need for the `Key` interface, but it limits 
the key type parameter to `comparable` types only, whereas `v8` allows `any` type, as long as it is wrapped as `Key`,
and exports itself as `string`.  

## Cache
This implementation contains a very basic cache that is intended only to be used for short-lived DataLoaders 
(i.e. DataLoaders that only exist for the life of a http request). You may use your own implementation if you want.

> it also has a `NoCache` type that implements the cache interface but all methods are noop. If you do not wish to cache anything.

## Examples
There are a few basic examples in the example folder.
