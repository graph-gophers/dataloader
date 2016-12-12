package dataloader

import (
	"fmt"
	"log"
	"reflect"
	"strconv"
	"sync"
	"testing"
)

///////////////////////////////////////////////////
// Tests
///////////////////////////////////////////////////
func TestLoader(t *testing.T) {
	t.Run("test Load method", func(t *testing.T) {
		t.Parallel()
		identityLoader, _ := IDLoader(0)
		future := identityLoader.Load("1")
		value, _ := future()
		if value != "1" {
			t.Error("load didn't return the right value")
		}
	})

	t.Run("test LoadMany method", func(t *testing.T) {
		t.Parallel()
		errorLoader, _ := ErrorLoader(0)
		future := errorLoader.LoadMany([]string{"1", "2", "3"})
		_, err := future()
		if len(err) != 3 {
			t.Error("loadmany didn't return right number of errors")
		}
	})

	t.Run("test LoadMany method", func(t *testing.T) {
		t.Parallel()
		identityLoader, _ := IDLoader(0)
		future := identityLoader.LoadMany([]string{"1", "2", "3"})
		results, _ := future()
		if results[0].(string) != "1" || results[1].(string) != "2" || results[2].(string) != "3" {
			t.Error("loadmany didn't return the right value")
		}
	})

	t.Run("batches many requests", func(t *testing.T) {
		t.Parallel()
		identityLoader, loadCalls := IDLoader(0)
		future1 := identityLoader.Load("1")
		future2 := identityLoader.Load("2")

		future1()
		future2()

		calls := *loadCalls
		inner := []string{"1", "2"}
		expected := [][]string{inner}
		if !reflect.DeepEqual(calls, expected) {
			t.Errorf("did not call batchFn in right order. Expected %#v, got %#v", expected, calls)
		}
	})

	t.Run("number of results matches number of keys", func(t *testing.T) {
		t.Parallel()
		faultyLoader, _ := FaultyLoader()

		n := 10
		reqs := []Thunk{}
		keys := []string{}
		for i := 0; i < n; i++ {
			key := strconv.Itoa(i)
			reqs = append(reqs, faultyLoader.Load(key))
			keys = append(keys, key)
		}

		for _, future := range reqs {
			future()
		}

		// TODO: expect to get some kind of warning
	})

	t.Run("responds to max batch size", func(t *testing.T) {
		t.Parallel()
		identityLoader, loadCalls := IDLoader(2)
		future1 := identityLoader.Load("1")
		future2 := identityLoader.Load("2")
		future3 := identityLoader.Load("3")

		future1()
		future2()
		future3()

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
		identityLoader, loadCalls := IDLoader(0)
		future1 := identityLoader.Load("1")
		future2 := identityLoader.Load("1")

		future1()
		future2()

		calls := *loadCalls
		inner := []string{"1"}
		expected := [][]string{inner}
		if !reflect.DeepEqual(calls, expected) {
			t.Errorf("did not respect max batch size. Expected %#v, got %#v", expected, calls)
		}
	})

	t.Run("allows primed cache", func(t *testing.T) {
		t.Parallel()
		identityLoader, loadCalls := IDLoader(0)
		identityLoader.Prime("A", "Cached")
		future1 := identityLoader.Load("1")
		future2 := identityLoader.Load("A")

		future1()
		value, _ := future2()

		calls := *loadCalls
		inner := []string{"1"}
		expected := [][]string{inner}
		if !reflect.DeepEqual(calls, expected) {
			t.Errorf("did not respect max batch size. Expected %#v, got %#v", expected, calls)
		}

		if value.(string) != "Cached" {
			t.Errorf("did not use primed cache value. Expected '%#v', got '%#v'", "Cached", value)
		}
	})

	t.Run("allows clear value in cache", func(t *testing.T) {
		t.Parallel()
		identityLoader, loadCalls := IDLoader(0)
		identityLoader.Prime("A", "Cached")
		identityLoader.Prime("B", "B")
		future1 := identityLoader.Load("1")
		future2 := identityLoader.Clear("A").Load("A")
		future3 := identityLoader.Load("B")

		future1()
		value, _ := future2()
		future3()

		calls := *loadCalls
		inner := []string{"1", "A"}
		expected := [][]string{inner}
		if !reflect.DeepEqual(calls, expected) {
			t.Errorf("did not respect max batch size. Expected %#v, got %#v", expected, calls)
		}

		if value.(string) != "A" {
			t.Errorf("did not use primed cache value. Expected '%#v', got '%#v'", "Cached", value)
		}
	})

	t.Run("allows clearAll values in cache", func(t *testing.T) {
		t.Parallel()
		identityLoader, loadCalls := IDLoader(0)
		identityLoader.Prime("A", "Cached")
		identityLoader.Prime("B", "B")

		identityLoader.ClearAll()

		future1 := identityLoader.Load("1")
		future2 := identityLoader.Load("A")
		future3 := identityLoader.Load("B")

		future1()
		future2()
		future3()

		calls := *loadCalls
		inner := []string{"1", "A", "B"}
		expected := [][]string{inner}
		if !reflect.DeepEqual(calls, expected) {
			t.Errorf("did not respect max batch size. Expected %#v, got %#v", expected, calls)
		}
	})

	t.Run("all methods on NoCache are Noops", func(t *testing.T) {
		t.Parallel()
		identityLoader, loadCalls := NoCacheLoader(0)
		identityLoader.Prime("A", "Cached")
		identityLoader.Prime("B", "B")

		identityLoader.ClearAll()

		future1 := identityLoader.Clear("1").Load("1")
		future2 := identityLoader.Load("A")
		future3 := identityLoader.Load("B")

		future1()
		future2()
		future3()

		calls := *loadCalls
		inner := []string{"1", "A", "B"}
		expected := [][]string{inner}
		if !reflect.DeepEqual(calls, expected) {
			t.Errorf("did not respect max batch size. Expected %#v, got %#v", expected, calls)
		}
	})

	t.Run("no cache does not cache anything", func(t *testing.T) {
		t.Parallel()
		identityLoader, loadCalls := NoCacheLoader(0)
		identityLoader.Prime("A", "Cached")
		identityLoader.Prime("B", "B")

		future1 := identityLoader.Load("1")
		future2 := identityLoader.Load("A")
		future3 := identityLoader.Load("B")

		future1()
		future2()
		future3()

		calls := *loadCalls
		inner := []string{"1", "A", "B"}
		expected := [][]string{inner}
		if !reflect.DeepEqual(calls, expected) {
			t.Errorf("did not respect max batch size. Expected %#v, got %#v", expected, calls)
		}
	})

}

// test helpers
func IDLoader(max int) (*Loader, *[][]string) {
	var mu sync.Mutex
	var loadCalls [][]string
	identityLoader := NewBatchedLoader(func(keys []string) []*Result {
		var results []*Result
		mu.Lock()
		loadCalls = append(loadCalls, keys)
		mu.Unlock()
		for _, key := range keys {
			results = append(results, &Result{key, nil})
		}
		return results
	}, WithBatchCapacity(max))
	return identityLoader, &loadCalls
}
func ErrorLoader(max int) (*Loader, *[][]string) {
	var mu sync.Mutex
	var loadCalls [][]string
	identityLoader := NewBatchedLoader(func(keys []string) []*Result {
		var results []*Result
		mu.Lock()
		loadCalls = append(loadCalls, keys)
		mu.Unlock()
		for _, key := range keys {
			results = append(results, &Result{key, fmt.Errorf("this is a test error")})
		}
		return results
	}, WithBatchCapacity(max))
	return identityLoader, &loadCalls
}
func BadLoader(max int) (*Loader, *[][]string) {
	var mu sync.Mutex
	var loadCalls [][]string
	identityLoader := NewBatchedLoader(func(keys []string) []*Result {
		var results []*Result
		mu.Lock()
		loadCalls = append(loadCalls, keys)
		mu.Unlock()
		results = append(results, &Result{keys[0], nil})
		return results
	}, WithBatchCapacity(max))
	return identityLoader, &loadCalls
}
func NoCacheLoader(max int) (*Loader, *[][]string) {
	var mu sync.Mutex
	var loadCalls [][]string
	cache := &NoCache{}
	identityLoader := NewBatchedLoader(func(keys []string) []*Result {
		var results []*Result
		mu.Lock()
		loadCalls = append(loadCalls, keys)
		mu.Unlock()
		for _, key := range keys {
			results = append(results, &Result{key, nil})
		}
		return results
	}, WithCache(cache), WithBatchCapacity(max))
	return identityLoader, &loadCalls
}

// FaultyLoader gives len(keys)-1 results.
func FaultyLoader() (*Loader, *[][]string) {
	var mu sync.Mutex
	var loadCalls [][]string

	loader := NewBatchedLoader(func(keys []string) []*Result {
		var results []*Result
		mu.Lock()
		loadCalls = append(loadCalls, keys)
		mu.Unlock()

		lastKeyIndex := len(keys) - 1
		for i, key := range keys {
			if i == lastKeyIndex {
				break
			}

			results = append(results, &Result{key, nil})
		}
		return results
	})

	return loader, &loadCalls
}

///////////////////////////////////////////////////
// Benchmarks
///////////////////////////////////////////////////
var a = &Avg{}

func batchIdentity(keys []string) (results []*Result) {
	a.Add(len(keys))
	for _, key := range keys {
		results = append(results, &Result{key, nil})
	}
	return
}

func BenchmarkLoader(b *testing.B) {
	UserLoader := NewBatchedLoader(batchIdentity)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		UserLoader.Load(strconv.Itoa(i))
	}
	log.Printf("avg: %f", a.Avg())
}

type Avg struct {
	total  float64
	length float64
}

func (a *Avg) Add(v int) {
	a.total += float64(v)
	a.length++
}

func (a *Avg) Avg() float64 {
	if a.total == 0 {
		return 0
	} else if a.length == 0 {
		return 0
	}
	return a.total / a.length
}
