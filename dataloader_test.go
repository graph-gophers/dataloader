package dataloader

import (
	"context"
	"errors"
	"fmt"
	"log"
	"reflect"
	"strconv"
	"sync"
	"testing"
	"time"
)

// /////////////////////////////////////////////////
// Tests
// /////////////////////////////////////////////////
func TestLoader(t *testing.T) {
	t.Run("test Load method", func(t *testing.T) {
		t.Parallel()
		identityLoader, _ := IDLoader[string](0)
		ctx := context.Background()
		future := identityLoader.Load(ctx, "1")
		value, err := future()
		if err != nil {
			t.Error(err.Error())
		}
		if value != "1" {
			t.Error("load didn't return the right value")
		}
	})

	t.Run("test thunk does not contain race conditions", func(t *testing.T) {
		t.Parallel()
		identityLoader, _ := IDLoader[string](0)
		ctx := context.Background()
		future := identityLoader.Load(ctx, "1")
		go future()
		go future()
	})

	t.Run("test Load Method Panic Safety", func(t *testing.T) {
		t.Parallel()
		defer func() {
			r := recover()
			if r != nil {
				t.Error("Panic Loader's panic should have been handled'")
			}
		}()
		panicLoader, _ := PanicLoader[string](0)
		ctx := context.Background()
		future := panicLoader.Load(ctx, "1")
		_, err := future()
		if err == nil || err.Error() != "Panic received in batch function: Programming error" {
			t.Error("Panic was not propagated as an error.")
		}
	})

	t.Run("test Load Method cache error", func(t *testing.T) {
		t.Parallel()
		errorCacheLoader, _ := ErrorCacheLoader[string](0)
		ctx := context.Background()
		futures := []Thunk[string]{}
		for i := 0; i < 2; i++ {
			futures = append(futures, errorCacheLoader.Load(ctx, strconv.Itoa(i)))
		}

		for _, f := range futures {
			_, err := f()
			if err == nil {
				t.Error("Error was not propagated")
			}
		}
		nextFuture := errorCacheLoader.Load(ctx, "1")
		_, err := nextFuture()

		// Normal errors should be cached.
		if err == nil {
			t.Error("Error from batch function was not cached")
		}
	})

	t.Run("test Load Method Panic Safety in multiple keys", func(t *testing.T) {
		t.Parallel()
		defer func() {
			r := recover()
			if r != nil {
				t.Error("Panic Loader's panic should have been handled'")
			}
		}()
		panicLoader, _ := PanicCacheLoader[string](0)
		futures := []Thunk[string]{}
		ctx := context.Background()
		for i := 0; i < 3; i++ {
			futures = append(futures, panicLoader.Load(ctx, strconv.Itoa(i)))
		}
		for _, f := range futures {
			_, err := f()
			if err == nil || err.Error() != "Panic received in batch function: Programming error" {
				t.Error("Panic was not propagated as an error.")
			}
		}

		futures = []Thunk[string]{}
		for i := 0; i < 3; i++ {
			futures = append(futures, panicLoader.Load(ctx, strconv.Itoa(1)))
		}

		for _, f := range futures {
			_, err := f()
			if err != nil {
				t.Error("Panic error from batch function was cached")
			}
		}
	})

	t.Run("test LoadMany returns errors", func(t *testing.T) {
		t.Parallel()
		errorLoader, _ := ErrorLoader[string](0)
		ctx := context.Background()
		future := errorLoader.LoadMany(ctx, []string{"1", "2", "3"})
		_, err := future()
		if len(err) != 3 {
			t.Error("LoadMany didn't return right number of errors")
		}
	})

	t.Run("test LoadMany returns len(errors) == len(keys)", func(t *testing.T) {
		t.Parallel()
		loader, _ := OneErrorLoader[string](3)
		ctx := context.Background()
		future := loader.LoadMany(ctx, []string{"1", "2", "3"})
		_, errs := future()
		if len(errs) != 3 {
			t.Errorf("LoadMany didn't return right number of errors (should match size of input)")
		}

		var errCount int = 0
		var nilCount int = 0
		for _, err := range errs {
			if err == nil {
				nilCount++
			} else {
				errCount++
			}
		}
		if errCount != 1 {
			t.Error("Expected an error on only one of the items loaded")
		}

		if nilCount != 2 {
			t.Error("Expected second and third errors to be nil")
		}
	})

	t.Run("test LoadMany returns nil []error when no errors occurred", func(t *testing.T) {
		t.Parallel()
		loader, _ := IDLoader[string](0)
		ctx := context.Background()
		_, err := loader.LoadMany(ctx, []string{"1", "2", "3"})()
		if err != nil {
			t.Errorf("Expected LoadMany() to return nil error slice when no errors occurred")
		}
	})

	t.Run("test thunkmany does not contain race conditions", func(t *testing.T) {
		t.Parallel()
		identityLoader, _ := IDLoader[string](0)
		ctx := context.Background()
		future := identityLoader.LoadMany(ctx, []string{"1", "2", "3"})
		go future()
		go future()
	})

	t.Run("test Load Many Method Panic Safety", func(t *testing.T) {
		t.Parallel()
		defer func() {
			r := recover()
			if r != nil {
				t.Error("Panic Loader's panic should have been handled'")
			}
		}()
		panicLoader, _ := PanicCacheLoader[string](0)
		ctx := context.Background()
		future := panicLoader.LoadMany(ctx, []string{"1", "2"})
		_, errs := future()
		if len(errs) < 2 || errs[0].Error() != "Panic received in batch function: Programming error" {
			t.Error("Panic was not propagated as an error.")
		}

		future = panicLoader.LoadMany(ctx, []string{"1"})
		_, errs = future()

		if len(errs) > 0 {
			t.Error("Panic error from batch function was cached")
		}

	})

	t.Run("test LoadMany method", func(t *testing.T) {
		t.Parallel()
		identityLoader, _ := IDLoader[string](0)
		ctx := context.Background()
		future := identityLoader.LoadMany(ctx, []string{"1", "2", "3"})
		results, _ := future()
		if results[0] != "1" || results[1] != "2" || results[2] != "3" {
			t.Error("loadmany didn't return the right value")
		}
	})

	t.Run("batches many requests", func(t *testing.T) {
		t.Parallel()
		identityLoader, loadCalls := IDLoader[string](0)
		ctx := context.Background()
		future1 := identityLoader.Load(ctx, "1")
		future2 := identityLoader.Load(ctx, "2")

		_, err := future1()
		if err != nil {
			t.Error(err.Error())
		}
		_, err = future2()
		if err != nil {
			t.Error(err.Error())
		}

		calls := *loadCalls
		inner := []string{"1", "2"}
		expected := [][]string{inner}
		if !reflect.DeepEqual(calls, expected) {
			t.Errorf("did not call batchFn in right order. Expected %#v, got %#v", expected, calls)
		}
	})

	t.Run("number of results matches number of keys", func(t *testing.T) {
		t.Parallel()
		faultyLoader, _ := FaultyLoader[string]()
		ctx := context.Background()

		n := 10
		reqs := []Thunk[string]{}
		var keys []string
		for i := 0; i < n; i++ {
			key := strconv.Itoa(i)
			reqs = append(reqs, faultyLoader.Load(ctx, key))
			keys = append(keys, key)
		}

		for _, future := range reqs {
			_, err := future()
			if err == nil {
				t.Error("if number of results doesn't match keys, all keys should contain error")
			}
		}

		// TODO: expect to get some kind of warning
	})

	t.Run("responds to max batch size", func(t *testing.T) {
		t.Parallel()
		identityLoader, loadCalls := IDLoader[string](2)
		ctx := context.Background()
		future1 := identityLoader.Load(ctx, "1")
		future2 := identityLoader.Load(ctx, "2")
		future3 := identityLoader.Load(ctx, "3")

		_, err := future1()
		if err != nil {
			t.Error(err.Error())
		}
		_, err = future2()
		if err != nil {
			t.Error(err.Error())
		}
		_, err = future3()
		if err != nil {
			t.Error(err.Error())
		}

		calls := *loadCalls
		inner1 := []string{"1", "2"}
		inner2 := []string{"3"}
		expected := [][]string{inner1, inner2}
		if !reflect.DeepEqual(calls, expected) {
			t.Errorf("did not respect max batch size. Expected %#v, got %#v", expected, calls)
		}
	})

	t.Run("caches repeated requests", func(t *testing.T) {
		t.Parallel()
		identityLoader, loadCalls := IDLoader[string](0)
		ctx := context.Background()
		future1 := identityLoader.Load(ctx, "1")
		future2 := identityLoader.Load(ctx, "1")

		_, err := future1()
		if err != nil {
			t.Error(err.Error())
		}
		_, err = future2()
		if err != nil {
			t.Error(err.Error())
		}

		calls := *loadCalls
		inner := []string{"1"}
		expected := [][]string{inner}
		if !reflect.DeepEqual(calls, expected) {
			t.Errorf("did not respect max batch size. Expected %#v, got %#v", expected, calls)
		}
	})

	t.Run("allows primed cache", func(t *testing.T) {
		t.Parallel()
		identityLoader, loadCalls := IDLoader[string](0)
		ctx := context.Background()
		identityLoader.Prime(ctx, "A", "Cached")
		future1 := identityLoader.Load(ctx, "1")
		future2 := identityLoader.Load(ctx, "A")

		_, err := future1()
		if err != nil {
			t.Error(err.Error())
		}
		value, err := future2()
		if err != nil {
			t.Error(err.Error())
		}

		calls := *loadCalls
		inner := []string{"1"}
		expected := [][]string{inner}
		if !reflect.DeepEqual(calls, expected) {
			t.Errorf("did not respect max batch size. Expected %#v, got %#v", expected, calls)
		}

		if value != "Cached" {
			t.Errorf("did not use primed cache value. Expected '%#v', got '%#v'", "Cached", value)
		}
	})

	t.Run("allows clear value in cache", func(t *testing.T) {
		t.Parallel()
		identityLoader, loadCalls := IDLoader[string](0)
		ctx := context.Background()
		identityLoader.Prime(ctx, "A", "Cached")
		identityLoader.Prime(ctx, "B", "B")
		future1 := identityLoader.Load(ctx, "1")
		future2 := identityLoader.Clear(ctx, "A").Load(ctx, "A")
		future3 := identityLoader.Load(ctx, "B")

		_, err := future1()
		if err != nil {
			t.Error(err.Error())
		}
		value, err := future2()
		if err != nil {
			t.Error(err.Error())
		}
		_, err = future3()
		if err != nil {
			t.Error(err.Error())
		}

		calls := *loadCalls
		inner := []string{"1", "A"}
		expected := [][]string{inner}
		if !reflect.DeepEqual(calls, expected) {
			t.Errorf("did not respect max batch size. Expected %#v, got %#v", expected, calls)
		}

		if value != "A" {
			t.Errorf("did not use primed cache value. Expected '%#v', got '%#v'", "Cached", value)
		}
	})

	t.Run("clears cache on batch with WithClearCacheOnBatch", func(t *testing.T) {
		t.Parallel()
		batchOnlyLoader, loadCalls := BatchOnlyLoader[string](0)
		ctx := context.Background()
		future1 := batchOnlyLoader.Load(ctx, "1")
		future2 := batchOnlyLoader.Load(ctx, "1")

		_, err := future1()
		if err != nil {
			t.Error(err.Error())
		}
		_, err = future2()
		if err != nil {
			t.Error(err.Error())
		}

		calls := *loadCalls
		inner := []string{"1"}
		expected := [][]string{inner}
		if !reflect.DeepEqual(calls, expected) {
			t.Errorf("did not batch queries. Expected %#v, got %#v", expected, calls)
		}

		if _, found := batchOnlyLoader.cache.Get(ctx, "1"); found {
			t.Errorf("did not clear cache after batch. Expected %#v, got %#v", false, found)
		}
	})

	t.Run("allows clearAll values in cache", func(t *testing.T) {
		t.Parallel()
		identityLoader, loadCalls := IDLoader[string](0)
		ctx := context.Background()
		identityLoader.Prime(ctx, "A", "Cached")
		identityLoader.Prime(ctx, "B", "B")

		identityLoader.ClearAll()

		future1 := identityLoader.Load(ctx, "1")
		future2 := identityLoader.Load(ctx, "A")
		future3 := identityLoader.Load(ctx, "B")

		_, err := future1()
		if err != nil {
			t.Error(err.Error())
		}
		_, err = future2()
		if err != nil {
			t.Error(err.Error())
		}
		_, err = future3()
		if err != nil {
			t.Error(err.Error())
		}

		calls := *loadCalls
		inner := []string{"1", "A", "B"}
		expected := [][]string{inner}
		if !reflect.DeepEqual(calls, expected) {
			t.Errorf("did not respect max batch size. Expected %#v, got %#v", expected, calls)
		}
	})

	t.Run("all methods on NoCache are Noops", func(t *testing.T) {
		t.Parallel()
		identityLoader, loadCalls := NoCacheLoader[string](0)
		ctx := context.Background()
		identityLoader.Prime(ctx, "A", "Cached")
		identityLoader.Prime(ctx, "B", "B")

		identityLoader.ClearAll()

		future1 := identityLoader.Clear(ctx, "1").Load(ctx, "1")
		future2 := identityLoader.Load(ctx, "A")
		future3 := identityLoader.Load(ctx, "B")

		_, err := future1()
		if err != nil {
			t.Error(err.Error())
		}
		_, err = future2()
		if err != nil {
			t.Error(err.Error())
		}
		_, err = future3()
		if err != nil {
			t.Error(err.Error())
		}

		calls := *loadCalls
		inner := []string{"1", "A", "B"}
		expected := [][]string{inner}
		if !reflect.DeepEqual(calls, expected) {
			t.Errorf("did not respect max batch size. Expected %#v, got %#v", expected, calls)
		}
	})

	t.Run("no cache does not cache anything", func(t *testing.T) {
		t.Parallel()
		identityLoader, loadCalls := NoCacheLoader[string](0)
		ctx := context.Background()
		identityLoader.Prime(ctx, "A", "Cached")
		identityLoader.Prime(ctx, "B", "B")

		future1 := identityLoader.Load(ctx, "1")
		future2 := identityLoader.Load(ctx, "A")
		future3 := identityLoader.Load(ctx, "B")

		_, err := future1()
		if err != nil {
			t.Error(err.Error())
		}
		_, err = future2()
		if err != nil {
			t.Error(err.Error())
		}
		_, err = future3()
		if err != nil {
			t.Error(err.Error())
		}

		calls := *loadCalls
		inner := []string{"1", "A", "B"}
		expected := [][]string{inner}
		if !reflect.DeepEqual(calls, expected) {
			t.Errorf("did not respect max batch size. Expected %#v, got %#v", expected, calls)
		}
	})

	t.Run("dataloader without timeout", func(t *testing.T) {
		t.Parallel()

		loader, loadCalls := TimeoutLoader[string](0, time.Millisecond*70)

		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*50)
		defer cancel()

		// ctx2 have second timeout, but will be used ctx with 50ms for batch
		ctx2, cancel2 := context.WithTimeout(context.Background(), time.Second)
		defer cancel2()

		future1 := loader.Load(ctx, "1")
		future2 := loader.Load(ctx2, "A")
		future3 := loader.Load(ctx, "B")

		_, err := future1()
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("wrong error code. Expected %#v, got %#v", context.DeadlineExceeded, err)
		}
		_, err = future2()
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("wrong error code. Expected %#v, got %#v", context.DeadlineExceeded, err)
		}
		_, err = future3()
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("wrong error code. Expected %#v, got %#v", context.DeadlineExceeded, err)
		}

		calls := *loadCalls
		inner := []string{"1", "A", "B"}
		expected := [][]string{inner}
		if !reflect.DeepEqual(calls, expected) {
			t.Errorf("did not respect max batch size. Expected %#v, got %#v", expected, calls)
		}

		time.Sleep(time.Millisecond * 10)

		ctx22, cancel22 := context.WithTimeout(context.Background(), time.Second)
		defer cancel22()

		// Because it was cached with an error
		future22 := loader.Load(ctx22, "A")
		_, err = future22()
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Errorf("wrong error code. Expected %#v, got %#v", context.DeadlineExceeded, err)
		}
	})

	t.Run("dataloader with timeout", func(t *testing.T) {
		t.Parallel()

		// setup detachContext with timeout 1s
		loader, loadCalls := TimeoutLoader[string](time.Second, time.Millisecond*70)

		ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*50)
		defer cancel()

		// ctx2 have second timeout, but will be used ctx with 50ms for batch
		ctx2, cancel2 := context.WithTimeout(context.Background(), time.Second)
		defer cancel2()

		future1 := loader.Load(ctx, "1")
		future2 := loader.Load(ctx2, "A")
		future3 := loader.Load(ctx, "B")

		_, err := future1()
		if err != nil {
			t.Error(err)
		}
		_, err = future2()
		if err != nil {
			t.Error(err)
		}
		_, err = future3()
		if err != nil {
			t.Error(err)
		}

		calls := *loadCalls
		inner := []string{"1", "A", "B"}
		expected := [][]string{inner}
		if !reflect.DeepEqual(calls, expected) {
			t.Errorf("did not respect max batch size. Expected %#v, got %#v", expected, calls)
		}
	})

	t.Run("datacache", func(t *testing.T) {
		t.Parallel()
		var mu sync.Mutex
		var calls [][]string
		identityLoader := DataCacheLoader[string, string](0, func(ctx context.Context, keys []string) []*Result[string] {
			result := make([]*Result[string], 0, len(keys))
			mu.Lock()
			defer mu.Unlock()
			calls = append(calls, keys)

			for _, key := range keys {
				result = append(result, &Result[string]{Data: key})
			}

			return result
		})

		ctx := context.Background()
		identityLoader.Prime(ctx, "A", "Cached")
		identityLoader.Prime(ctx, "B", "B")

		future1 := identityLoader.Load(ctx, "1")
		future2 := identityLoader.Load(ctx, "A")
		future3 := identityLoader.Load(ctx, "B")

		_, err := future1()
		if err != nil {
			t.Error(err.Error())
		}
		_, err = future2()
		if err != nil {
			t.Error(err.Error())
		}
		_, err = future3()
		if err != nil {
			t.Error(err.Error())
		}

		// waiting first batch end, add data added to datacache
		time.Sleep(time.Millisecond * 50)
		_, err = future2()
		if err != nil {
			t.Error(err.Error())
		}
		_, err = future3()
		if err != nil {
			t.Error(err.Error())
		}
		future4 := identityLoader.Load(ctx, "C")
		_, err = future4()
		if err != nil {
			t.Error(err.Error())
		}

		inner := []string{"1", "A", "B"}
		inner2 := []string{"C"}
		expected := [][]string{inner, inner2}
		if !reflect.DeepEqual(calls, expected) {
			t.Errorf("did not respect max batch size. Expected %#v, got %#v", expected, calls)
		}
	})

	t.Run("batches many requests with datacache and double keys", func(t *testing.T) {
		t.Parallel()
		identityLoader, loadCalls := IDLoaderDataCache[string](0)
		ctx := context.Background()
		future1 := identityLoader.Load(ctx, "1")
		future2 := identityLoader.Load(ctx, "2")
		future3 := identityLoader.Load(ctx, "1")

		val1, err := future1()
		if err != nil {
			t.Error(err.Error())
		}
		if val1 != "1" {
			t.Errorf("future1 expected %#v, got %#v", "1", val1)
		}

		val2, err := future2()
		if err != nil {
			t.Error(err.Error())
		}
		if val2 != "2" {
			t.Errorf("future2 expected %#v, got %#v", "2", val2)
		}

		val3, err := future3()
		if err != nil {
			t.Error(err.Error())
		}
		if val1 != "1" {
			t.Errorf("future3 expected %#v, got %#v", "1", val3)
		}

		calls := *loadCalls
		inner := []string{"1", "2"} // batch double keys
		expected := [][]string{inner}
		if !reflect.DeepEqual(calls, expected) {
			t.Errorf("did not call batchFn in right order. Expected %#v, got %#v", expected, calls)
		}
	})

	t.Run("struct datacache", func(t *testing.T) {
		t.Parallel()
		var mu sync.Mutex
		var calls [][]int
		identityLoader := DataCacheLoader[key[userKey], string](0, func(ctx context.Context, keys []key[userKey]) []*Result[string] {
			result := make([]*Result[string], 0, len(keys))
			loadCalls := make([]int, 0, len(keys))
			for _, key := range keys {
				loadCalls = append(loadCalls, key.Raw().ID)

				result = append(result, &Result[string]{Data: fmt.Sprintf("%v", key.Raw().ID)})
			}

			mu.Lock()
			defer mu.Unlock()
			calls = append(calls, loadCalls)

			return result
		})
		ctx := context.Background()
		identityLoader.Prime(ctx, ContextKey(ctx, userKey{ID: 1}), "Cached")
		identityLoader.Prime(ctx, ContextKey(ctx, userKey{ID: 2}), "B")

		future1 := identityLoader.Load(ctx, ContextKey(ctx, userKey{ID: 1}))
		future2 := identityLoader.Load(ctx, ContextKey(ctx, userKey{ID: 2}))
		future3 := identityLoader.Load(ctx, ContextKey(ctx, userKey{ID: 3}))

		_, err := future1()
		if err != nil {
			t.Error(err.Error())
		}
		_, err = future2()
		if err != nil {
			t.Error(err.Error())
		}
		_, err = future3()
		if err != nil {
			t.Error(err.Error())
		}

		// waiting first batch end, add data added to datacache
		time.Sleep(time.Millisecond * 50)
		_, err = future2()
		if err != nil {
			t.Error(err.Error())
		}
		_, err = future3()
		if err != nil {
			t.Error(err.Error())
		}
		future4 := identityLoader.Load(ctx, ContextKey(ctx, userKey{ID: 4}))
		_, err = future4()
		if err != nil {
			t.Error(err.Error())
		}

		inner := []int{1, 2, 3}
		inner2 := []int{4}
		expected := [][]int{inner, inner2}
		if !reflect.DeepEqual(calls, expected) {
			t.Errorf("did not respect max batch size. Expected %#v, got %#v", expected, calls)
		}
	})
}

type key[K comparable] struct {
	root K
	ctx  context.Context
}

func ContextKey[K comparable](ctx context.Context, k K) key[K] {
	return key[K]{root: k, ctx: ctx}
}

func (k *key[K]) Raw() K {
	return k.root
}

func (k *key[K]) Context() context.Context {
	return k.ctx
}

type userKey struct {
	ID int
}

// test helpers
func IDLoader[K comparable](max int) (*Loader[K, K], *[][]K) {
	var mu sync.Mutex
	var loadCalls [][]K
	identityLoader := NewBatchedLoader(func(_ context.Context, keys []K) []*Result[K] {
		var results []*Result[K]
		mu.Lock()
		loadCalls = append(loadCalls, keys)
		mu.Unlock()
		for _, key := range keys {
			results = append(results, &Result[K]{key, nil})
		}
		return results
	}, WithBatchCapacity[K, K](max))
	return identityLoader, &loadCalls
}
func BatchOnlyLoader[K comparable](max int) (*Loader[K, K], *[][]K) {
	var mu sync.Mutex
	var loadCalls [][]K
	identityLoader := NewBatchedLoader(func(_ context.Context, keys []K) []*Result[K] {
		var results []*Result[K]
		mu.Lock()
		loadCalls = append(loadCalls, keys)
		mu.Unlock()
		for _, key := range keys {
			results = append(results, &Result[K]{key, nil})
		}
		return results
	}, WithBatchCapacity[K, K](max), WithClearCacheOnBatch[K, K]())
	return identityLoader, &loadCalls
}
func ErrorLoader[K comparable](max int) (*Loader[K, K], *[][]K) {
	var mu sync.Mutex
	var loadCalls [][]K
	identityLoader := NewBatchedLoader(func(_ context.Context, keys []K) []*Result[K] {
		var results []*Result[K]
		mu.Lock()
		loadCalls = append(loadCalls, keys)
		mu.Unlock()
		for _, key := range keys {
			results = append(results, &Result[K]{key, fmt.Errorf("this is a test error")})
		}
		return results
	}, WithBatchCapacity[K, K](max))
	return identityLoader, &loadCalls
}
func OneErrorLoader[K comparable](max int) (*Loader[K, K], *[][]K) {
	var mu sync.Mutex
	var loadCalls [][]K
	identityLoader := NewBatchedLoader(func(_ context.Context, keys []K) []*Result[K] {
		results := make([]*Result[K], max)
		mu.Lock()
		loadCalls = append(loadCalls, keys)
		mu.Unlock()
		for i := range keys {
			var err error
			if i == 0 {
				err = errors.New("always error on the first key")
			}
			results[i] = &Result[K]{keys[i], err}
		}
		return results
	}, WithBatchCapacity[K, K](max))
	return identityLoader, &loadCalls
}
func PanicLoader[K comparable](max int) (*Loader[K, K], *[][]K) {
	var loadCalls [][]K
	panicLoader := NewBatchedLoader(func(_ context.Context, keys []K) []*Result[K] {
		panic("Programming error")
	}, WithBatchCapacity[K, K](max), withSilentLogger[K, K]())
	return panicLoader, &loadCalls
}

func PanicCacheLoader[K comparable](max int) (*Loader[K, K], *[][]K) {
	var loadCalls [][]K
	panicCacheLoader := NewBatchedLoader(func(_ context.Context, keys []K) []*Result[K] {
		if len(keys) > 1 {
			panic("Programming error")
		}

		returnResult := make([]*Result[K], len(keys))
		for idx := range returnResult {
			returnResult[idx] = &Result[K]{
				keys[0],
				nil,
			}
		}

		return returnResult

	}, WithBatchCapacity[K, K](max), withSilentLogger[K, K]())
	return panicCacheLoader, &loadCalls
}

func ErrorCacheLoader[K comparable](max int) (*Loader[K, K], *[][]K) {
	var loadCalls [][]K
	errorCacheLoader := NewBatchedLoader(func(_ context.Context, keys []K) []*Result[K] {
		if len(keys) > 1 {
			var results []*Result[K]
			for _, key := range keys {
				results = append(results, &Result[K]{key, fmt.Errorf("this is a test error")})
			}
			return results
		}

		returnResult := make([]*Result[K], len(keys))
		for idx := range returnResult {
			returnResult[idx] = &Result[K]{
				keys[0],
				nil,
			}
		}

		return returnResult

	}, WithBatchCapacity[K, K](max), withSilentLogger[K, K]())
	return errorCacheLoader, &loadCalls
}

func BadLoader[K comparable](max int) (*Loader[K, K], *[][]K) {
	var mu sync.Mutex
	var loadCalls [][]K
	identityLoader := NewBatchedLoader(func(_ context.Context, keys []K) []*Result[K] {
		var results []*Result[K]
		mu.Lock()
		loadCalls = append(loadCalls, keys)
		mu.Unlock()
		results = append(results, &Result[K]{keys[0], nil})
		return results
	}, WithBatchCapacity[K, K](max))
	return identityLoader, &loadCalls
}

func NoCacheLoader[K comparable](max int) (*Loader[K, K], *[][]K) {
	var mu sync.Mutex
	var loadCalls [][]K
	cache := &NoCache[K, K]{}
	identityLoader := NewBatchedLoader(func(_ context.Context, keys []K) []*Result[K] {
		var results []*Result[K]
		mu.Lock()
		loadCalls = append(loadCalls, keys)
		mu.Unlock()
		for _, key := range keys {
			results = append(results, &Result[K]{key, nil})
		}
		return results
	}, WithCache[K, K](cache), WithBatchCapacity[K, K](max))
	return identityLoader, &loadCalls
}

// test helpers
func IDLoaderDataCache[K comparable](max int) (*Loader[K, K], *[][]K) {
	var mu sync.Mutex
	var loadCalls [][]K
	identityLoader := DataCacheLoader(max, func(_ context.Context, keys []K) []*Result[K] {
		var results []*Result[K]
		mu.Lock()
		loadCalls = append(loadCalls, keys)
		mu.Unlock()
		for _, key := range keys {
			results = append(results, &Result[K]{key, nil})
		}
		return results
	})
	return identityLoader, &loadCalls
}

func DataCacheLoader[K comparable, V any](max int, fn BatchFunc[K, V]) *Loader[K, V] {
	cache := &NoCache[K, V]{}

	dcacheData := make(map[K]V, max)
	var dcachemu sync.Mutex
	datacache := &dcache[K, V]{set: func(ctx context.Context, key K, value V) {
		dcachemu.Lock()
		defer dcachemu.Unlock()
		dcacheData[key] = value
	}, get: func(ctx context.Context, key K) (V, bool) {
		dcachemu.Lock()
		defer dcachemu.Unlock()

		data, ok := dcacheData[key]
		return data, ok
	}, del: func(ctx context.Context, k K) bool {
		dcachemu.Lock()
		defer dcachemu.Unlock()

		delete(dcacheData, k)
		return true
	}, clear: func() {
		dcachemu.Lock()
		defer dcachemu.Unlock()

		dcacheData = make(map[K]V)
	}}

	identityLoader := NewBatchedLoader(fn, WithCache[K, V](cache), WithBatchCapacity[K, V](max), WithDataCache[K, V](datacache))
	return identityLoader
}

func TimeoutLoader[K comparable](timeout time.Duration, sleepTimeout time.Duration) (*Loader[K, K], *[][]K) {
	var mu sync.Mutex
	var loadCalls [][]K

	fn := func(ctx context.Context, keys []K) []*Result[K] {
		time.Sleep(sleepTimeout)

		var results []*Result[K]
		mu.Lock()
		loadCalls = append(loadCalls, keys)
		mu.Unlock()

		for _, key := range keys {
			if ctx.Err() == nil {
				results = append(results, &Result[K]{key, nil})
			} else {
				results = append(results, &Result[K]{Error: ctx.Err()})
			}
		}
		return results
	}

	identityLoader := NewBatchedLoader(fn, WithTimeout[K, K](timeout))
	return identityLoader, &loadCalls
}

// FaultyLoader gives len(keys)-1 results.
func FaultyLoader[K comparable]() (*Loader[K, K], *[][]K) {
	var mu sync.Mutex
	var loadCalls [][]K

	loader := NewBatchedLoader(func(_ context.Context, keys []K) []*Result[K] {
		var results []*Result[K]
		mu.Lock()
		loadCalls = append(loadCalls, keys)
		mu.Unlock()

		lastKeyIndex := len(keys) - 1
		for i, key := range keys {
			if i == lastKeyIndex {
				break
			}

			results = append(results, &Result[K]{key, nil})
		}
		return results
	})

	return loader, &loadCalls
}

// DataCache
type dcache[K comparable, V any] struct {
	get   func(context.Context, K) (V, bool)
	set   func(context.Context, K, V)
	del   func(context.Context, K) bool
	clear func()
}

func (d *dcache[K, V]) Get(ctx context.Context, key K) (V, bool) {
	return d.get(ctx, key)
}

func (d *dcache[K, V]) Set(ctx context.Context, key K, value V) {
	d.set(ctx, key, value)
}

func (d *dcache[K, V]) Delete(ctx context.Context, key K) bool {
	return d.del(ctx, key)
}

func (d *dcache[K, V]) Clear() {
	d.clear()
}

// /////////////////////////////////////////////////
// Benchmarks
// /////////////////////////////////////////////////
var a = &Avg{}

func batchIdentity[K comparable](_ context.Context, keys []K) (results []*Result[K]) {
	a.Add(len(keys))
	for _, key := range keys {
		results = append(results, &Result[K]{key, nil})
	}
	return
}

var _ctx = context.Background()

func BenchmarkLoader(b *testing.B) {
	UserLoader := NewBatchedLoader(batchIdentity[string])
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		UserLoader.Load(_ctx, (strconv.Itoa(i)))
	}
	log.Printf("avg: %f", a.Avg())
}

type Avg struct {
	total  float64
	length float64
	lock   sync.RWMutex
}

func (a *Avg) Add(v int) {
	a.lock.Lock()
	a.total += float64(v)
	a.length++
	a.lock.Unlock()
}

func (a *Avg) Avg() float64 {
	a.lock.RLock()
	defer a.lock.RUnlock()
	if a.total == 0 {
		return 0
	} else if a.length == 0 {
		return 0
	}
	return a.total / a.length
}
