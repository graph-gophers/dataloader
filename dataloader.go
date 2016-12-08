// Package dataloader is an implimentation of facebook's dataloader in go.
// See https://github.com/facebook/dataloader for more information
package dataloader

import (
	"sync"
	"time"
)

// the buffer amount fot the input channel.
// I wish we had unbounded channels
var inputCap int = 1000

// Interface is a `DataLoader` Interface which defines a public API for loading data from a particular
// data back-end with unique keys such as the `id` column of a SQL table or
// document name in a MongoDB database, given a batch loading function.
//
// Each `DataLoader` instance should contain a unique memoized cache. Use caution when
// used in long-lived applications or those which serve many users with
// different access permissions and consider creating a new instance per
// web request.
type Interface interface {
	Load(string) Thunk
	LoadMany([]string) ThunkMany
	Clear(string) Interface
	ClearAll() Interface
	Prime(key string, value interface{}) Interface
}

// BatchFunc is a function, which when given a slice of keys (string), returns an slice of `results`.
// It's important that the length of the input keys matches the length of the ouput results.
//
// The keys passed to this function are guaranteed to be unique
type BatchFunc func([]string) []*Result

// Result is the data structure that a BatchFunc returns.
// It contains the resolved data, and any errors that may have occured while fetching the data.
type Result struct {
	Data  interface{}
	Error error
}

// ResultMany is used by the loadMany method. It contains a list of resolved data and a list of erros // if any occured.
// Errors will contain the index of the value that errored
type ResultMany struct {
	Data  []interface{}
	Error []error
}

// Loader implements the dataloader.Interface.
type Loader struct {
	// the batch function to be used by this loader
	batchFn BatchFunc

	// the maximum batch size. Set to 0 if you want it to be unbounded.
	cap int

	// the internal cache. This packages contains a basic cache implementation but any custom cache
	// implementation could be used as long as it implements the `Cache` interface.
	cacheLock sync.Mutex
	cache     Cache

	// used to close the input channel early
	forceStartBatch chan bool

	// connt of queued up items
	countLock sync.Mutex
	count     int

	// internal channel that is used to batch items
	inputLock sync.RWMutex
	input     chan *batchRequest
	batching  bool
}

// Thunk is a function that will block until the value (*Result) it contins is resolved.
// After the value it contians is resolved, this function will return the result.
// This function can be called many times, much like a Promise is other languages.
// The value will only need to be resolved once so subsequent calls will return immediately.
type Thunk func() *Result

// ThunkMany is much like the Thunk func type but it contains a list of results.
type ThunkMany func() *ResultMany

// type used to on input channel
type batchRequest struct {
	key     string
	channel chan *Result
}

// this help match the error to the key of a specific index
type resultError struct {
	error
	index int
}

// NewBatchedLoader constructs a new Loader with given options
func NewBatchedLoader(batchFn BatchFunc, cache Cache, cap int) *Loader {
	return &Loader{
		batchFn:         batchFn,
		cache:           cache,
		cap:             cap,
		forceStartBatch: make(chan bool),
		count:           0,
		input:           make(chan *batchRequest, inputCap),
	}
}

// Load load/resolves the given key, returning a channel that will contain the value and error
func (l *Loader) Load(key string) Thunk {
	var value *Result
	c := make(chan *Result, 1)
	cond := sync.NewCond(&sync.Mutex{})

	// lock to prevent duplicate keys coming in before item has been added to cache.
	l.cacheLock.Lock()
	if v, ok := l.cache.Get(key); ok {
		defer l.cacheLock.Unlock()
		return v
	}

	thunk := func() *Result {
		if value == nil {
			cond.L.Lock()
			cond.Wait()
			cond.L.Unlock()
		}
		return value
	}

	l.cache.Set(key, thunk)
	l.cacheLock.Unlock()

	// this is sent to batch fn. It contains the key and the channel to return the
	// the result on
	req := &batchRequest{key, c}

	// start the batch window if it hasn't already started.
	if !l.batching {
		l.inputLock.Lock()
		l.batching = true
		l.inputLock.Unlock()
		go l.batch()
	}

	// this lock prevents sending on the channel at the same time that it is being closed.
	l.inputLock.RLock()
	l.input <- req
	l.inputLock.RUnlock()

	// if we need to keep track of the count (max batch), then do so.
	if l.cap > 0 {
		l.countLock.Lock()
		l.count++
		l.countLock.Unlock()

		// if we hit our limit, force the batch to start
		if l.count == l.cap {
			l.forceStartBatch <- true
		}
	}

	defer func() {
		go func() {
			value = <-c
			cond.Broadcast()
		}()
	}()

	return thunk
}

// LoadMany loads mulitiple keys, returning a thunk (type: ThunkMany) that will resolve the keys passed in.
func (l *Loader) LoadMany(keys []string) ThunkMany {
	length := len(keys)
	data := make([]interface{}, length)
	errors := make([]error, 0, length)
	c := make(chan *ResultMany, 1)
	cond := sync.NewCond(&sync.Mutex{})
	wg := sync.WaitGroup{}

	wg.Add(length)
	for i := range keys {
		go func(i int) {
			defer wg.Done()
			thunk := l.Load(keys[i])
			result := thunk()
			if result.Error != nil {
				errors = append(errors, resultError{result.Error, i})
			}
			data[i] = result.Data
		}(i)
	}

	go func() {
		wg.Wait()
		c <- &ResultMany{data, errors}
		close(c)
	}()

	var value *ResultMany

	thunkMany := func() *ResultMany {
		if value == nil {
			cond.L.Lock()
			cond.Wait()
			cond.L.Unlock()
		}
		return value
	}

	defer func() {
		go func() {
			value = <-c
			cond.Broadcast()
		}()
	}()

	return thunkMany
}

// Clear clears the value at `key` from the cache, it it exsits. Returs self for method chaining
func (l *Loader) Clear(key string) Interface {
	l.cache.Delete(key)
	return l
}

// ClearAll clears the entire cache. To be used when some event results in unknown invalidations.
// Returns self for method chaining.
func (l *Loader) ClearAll() Interface {
	l.cache.Clear()
	return l
}

// Prime adds the provided key and value to the cache. If the key already exists, no change is made.
// Returns self for method chaining
func (l *Loader) Prime(key string, value interface{}) Interface {
	if _, ok := l.cache.Get(key); !ok {
		future := func() *Result {
			return &Result{
				Data:  value,
				Error: nil,
			}
		}
		l.cache.Set(key, future)
	}
	return l
}

// execuite the batch of all items in queue
func (l *Loader) batch() {
	var keys []string
	var reqs []*batchRequest

	go l.sleeper()

	for item := range l.input {
		keys = append(keys, item.key)
		reqs = append(reqs, item)
	}

	items := l.batchFn(keys)

	for i, req := range reqs {
		req.channel <- items[i]
		close(req.channel)
	}
}

// wait the appropriate amount of time for next batch
func (l *Loader) sleeper() {
	select {
	// used by batch to close early. usually triggered by max batch size
	case <-l.forceStartBatch:
	// this will move this goroutine to the back of the callstack?
	case <-time.After(10 * time.Millisecond):
	}

	// reset
	l.inputLock.Lock()
	close(l.input)
	l.input = make(chan *batchRequest, inputCap)
	l.batching = false
	l.inputLock.Unlock()
}
