// dataloader is an implimentation of facebook's dataloader in go.
// See https://github.com/facebook/dataloader for more information
package dataloader

import (
	"fmt"
	"sync"
	"time"
)

const inputCap int = 10

// A `DataLoader` Interface defines a public API for loading data from a particular
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

// BatchFunc is a Function, which when given a Slice of keys (string), returns an slice of `results`.
// It's important that the length of the input keys matches the length of the ouput results.
type BatchFunc func([]string) []*Result

// Result is the data structure that is primarily used by the BatchFunc.
// It contains the resolved data, and any errors that may have occured while fetching the data.
type Result struct {
	Data  interface{}
	Error error
}

// ResultMany is used by the loadMany method. It contains a list of resolved data and a list of erros // if any occured.
type ResultMany struct {
	Data  []interface{}
	Error []error
}

// Implements the dataloader.Interface.
type Loader struct {
	// the batch function to be used by this loader
	batchFn BatchFunc
	// the maximum batch size. Set to 0 if you want it to be unbounded.
	cap int
	// the internal cache. This packages contains a basic cache implementation but any custom cache
	// implementation could be used as long as it implements the `Cache` interface.
	cache Cache
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

// Future is a function that will block until the value it contins is resolved.
// After the value it contians is resolved, this function will return the result.
// This function can be called many times, much like a Promise is other languages.
type Thunk func() *Result

// FutureMany is much like the Future func type but it contains a list of results.
type ThunkMany func() *ResultMany

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
	c := make(chan *Result, l.cap)
	cond := sync.NewCond(&sync.Mutex{})
	if v, ok := l.cache.Get(key); ok {
		return v
	}

	var value *Result

	thunk := func() *Result {
		cond.L.Lock()
		defer cond.L.Unlock()
		if value == nil {
			cond.Wait()
		}
		return value
	}

	l.cache.Set(key, thunk)
	req := &batchRequest{key, c}

	if !l.batching {
		l.inputLock.Lock()
		l.batching = true
		l.inputLock.Unlock()
		go l.batch()
	}

	l.inputLock.RLock()
	l.input <- req
	l.inputLock.RUnlock()

	l.countLock.Lock()
	l.count = l.count + 1
	l.countLock.Unlock()

	if l.cap > 0 && l.count >= l.cap {
		l.forceStartBatch <- true
	}

	defer func() {
		go func() {
			select {
			case result := <-c:
				value = result
				cond.Broadcast()
			}
		}()
	}()

	return thunk
}

// LoadMany loads mulitiple keys, returning a channel that will resolve to an array of values and an error
func (l *Loader) LoadMany(keys []string) ThunkMany {
	out := make(chan *ResultMany, 1)
	c := make(chan *result, len(keys))

	for i := range keys {
		go func(i int) {
			getter := l.Load(keys[i])
			c <- &result{getter(), i}
		}(i)
	}

	go func() {
		defer close(out)
		var i int = 1
		outputData := make([]interface{}, len(keys), len(keys))
		outputErrors := make([]error, 0, len(keys))
		defer func() {
			if r := recover(); r != nil {
				fmt.Println("Recovered in f", r)
			}
		}()
		for result := range c {
			outputData[result.index] = result.Data
			if result.Error != nil {
				outputErrors = append(outputErrors, resultError{result.Error, result.index})
			}
			if i == len(keys) {
				close(c)
			}
			i++
		}
		out <- &ResultMany{outputData, outputErrors}
	}()

	var value struct {
		value *ResultMany
		lock  sync.Mutex
	}

	thunkMany := func() *ResultMany {
		if v, ok := <-out; ok {
			value.lock.Lock()
			value.value = v
			value.lock.Unlock()
		}

		return value.value
	}

	return thunkMany
}

// Clear clears the value at `key` from the cache, it it exsits. Returs self for method chaining
func (l *Loader) Clear(key string) Interface {
	l.cache.Delete(key)
	return l
}

// ClearAll clears the entire cache. To be used when some event results in unknown invalidations.
// Returs self for method chaining.
func (l *Loader) ClearAll() Interface {
	l.cache.Clear()
	return l
}

// Adds the provided key and value to the cache. If the key already exists, no change is made.
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

// type used to on input channel
type batchRequest struct {
	key     string
	channel chan *Result
}

// this help ensure that data return is in same order as data recieved
type result struct {
	*Result
	index int
}

// this help match the error to the key of a specific index
type resultError struct {
	error
	index int
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

	if len(items) != len(reqs) {
		for _, req := range reqs {
			req.channel <- &Result{
				Error: fmt.Errorf("length of keys must match length of responses"),
				Data:  nil,
			}
			close(req.channel)
		}
	} else {
		for i, req := range reqs {
			req.channel <- items[i]
			close(req.channel)
		}
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

	l.inputLock.Lock()
	close(l.input)
	l.input = make(chan *batchRequest, inputCap)
	l.batching = false
	l.inputLock.Unlock()
}
