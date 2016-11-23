// dataloader is an implimentation of facebook's dataloader in go.
// See https://github.com/facebook/dataloader for more information
package dataloader

import (
	"fmt"
	"log"
	"sync"
	"time"
)

// A `DataLoader` Interface defines a public API for loading data from a particular
// data back-end with unique keys such as the `id` column of a SQL table or
// document name in a MongoDB database, given a batch loading function.
//
// Each `DataLoader` instance should contain a unique memoized cache. Use caution when
// used in long-lived applications or those which serve many users with
// different access permissions and consider creating a new instance per
// web request.
type Interface interface {
	Load(string) <-chan *Result
	LoadMany([]string) <-chan *ResultMany
	Clear(string) Interface
	ClearAll() Interface
	Prime(key string, value interface{}) Interface
}

// BatchFunc is a Function, which when given an Array of keys (string), returns an array of `results`.
// It's important the the length of the input keys must match the length of
// the ouput results.
type BatchFunc func([]string) []Result

// Result is the data structure that is primarily used by the BatchFunc. It contains the resolved data,
// and any errors that may have occured while fetching the data.
type Result struct {
	Data  interface{}
	Error error
}

// ResultMany is used by the loadMany method. It contains a list of resolved data and a list of erros
// if any occured.
type ResultMany struct {
	Data  []interface{}
	Error []error
}

// Implements the dataloader.Interface.
type Loader struct {
	// the batch function to be used by this loader
	batchFn BatchFunc
	// the interval on how often to batch requests. The javascript implementation uses the event loop
	// to determine how often to batch, that concept doens't apply well to Golang so I would recommend
	// using a value <= 16 milliseconds to mimick that functionality
	timeout time.Duration
	// the maximum batch size. Set to 0 if you want it to be unbounded.
	cap int
	// the internal cache. This packages contains a basic cache implementation but any custom cache
	// implementation could be used as long as it implements the `Cache` interface.
	cache Cache

	// internal channel that is used to batch items
	input chan *batchRequest
	// input mutex
	mu sync.Mutex
}

// type used to on input channel
type batchRequest struct {
	key     string
	channel chan *Result
}

// NewBatchedLoader constructs a new Loader with given options
func NewBatchedLoader(batchFn BatchFunc, timeout time.Duration, cache Cache, cap int) *Loader {
	return &Loader{
		batchFn: batchFn,
		cache:   cache,
		timeout: timeout,
		cap:     cap,
	}
}

// Load load/resolves the given key, returning a channel that will contain the value and error
func (l *Loader) Load(key string) <-chan *Result {
	c := make(chan *Result, 1)
	if v, ok := l.cache.Get(key); ok {
		defer func() {
			c <- &Result{v, nil}
			close(c)
		}()
		return c
	}
	req := &batchRequest{key, c}
	l.addToQueue(req)
	return c
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

// LoadMany loads mulitiple keys, returning a channel that will resolve to an array of values and an error
func (l *Loader) LoadMany(keys []string) <-chan *ResultMany {
	out := make(chan *ResultMany, 1)
	c := make(chan *result, len(keys))

	for i := range keys {
		go func(i int) {
			future := l.Load(keys[i])
			value := <-future
			c <- &result{value, i}
		}(i)
	}

	go func() {
		defer close(out)
		var i int = 1
		outputData := make([]interface{}, len(keys), len(keys))
		outputErrors := make([]error, 0, len(keys))
		for result := range c {
			log.Printf("res: %#v", result.Data)
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

	return out
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
		l.cache.Set(key, value)
	}
	return l
}

// queues item to be batched
func (l *Loader) addToQueue(req *batchRequest) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.input == nil {
		l.input = make(chan *batchRequest, l.cap)
		go l.batch()
	}

	l.input <- req
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

	c := make(chan []Result)
	go func(channel chan []Result) {
		channel <- l.batchFn(keys)
	}(c)
	items := <-c

	if len(items) != len(reqs) {
		panic(fmt.Errorf("length of keys must match length of responses"))
	}

	for i, req := range reqs {
		l.cache.Set(req.key, items[i].Data)
		req.channel <- &items[i]
		close(req.channel)
	}
}

// wait the appropriate amount of time for next batch
func (l *Loader) sleeper() {
	time.Sleep(l.timeout)

	l.mu.Lock()
	defer l.mu.Unlock()

	close(l.input)
	l.input = nil
}
