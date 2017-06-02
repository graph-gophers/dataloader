# DataLoader
[![GoDoc](https://godoc.org/github.com/nicksrandall/dataloader?status.svg)](https://godoc.org/github.com/nicksrandall/dataloader)
[![Build Status](https://travis-ci.org/nicksrandall/dataloader.svg?branch=master)](https://travis-ci.org/nicksrandall/dataloader)
[![codecov](https://codecov.io/gh/nicksrandall/dataloader/branch/master/graph/badge.svg)](https://codecov.io/gh/nicksrandall/dataloader)

This is an implementation of [Facebook's DataLoader](https://github.com/facebook/dataloader) in Golang.

## Status
This project is a work in progress. Feedback is encouraged.

## Install
`go get -u gopkg.in/nicksrandall/dataloader.v2`

## Usage
```go
// setup batch function
batchFn := func(ctx context.Context, keys []string) []dataloader.Result {
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
thunk := loader.Load(ctx.TODO(), "key1")
result, err := thunk()
if err != nil {
  // handle data error
}

log.Printf("value: %#v", result)
```

## Upgrade from v1
The only difference between v1 and v2 is that we added use of [context](https://golang.org/pkg/context).

```diff
- loader.Load(key string) Thunk
+ loader.Load(ctx context.Context, key string) Thunk
- loader.LoadMany(keys []string) ThunkMany
+ loader.LoadMany(ctx context.Context, keys []string) ThunkMany
```

```diff
- type BatchFunc func([]string) []*Result
+ type BatchFunc func(context.Context, []string) []*Result
```

### Don't need/want to use context?
You're welcome to install the v1 version of this library.

## Cache
This implementation contains a very basic cache that is intended only to be used for short lived DataLoaders (i.e. DataLoaders that ony exsist for the life of an http request). You may use your own implementation if you want.

> it also has a `NoCache` type that implements the cache interface but all methods are noop. If you do not wish to cache anyting.

## Examples
There are a few basic examples in the example folder.
