package dataloader

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"
)

var ts *httptest.Server

var data map[string]string
var lock sync.Mutex

func init() {
	data = map[string]string{
		"1":  "Audi",
		"2":  "Bently",
		"3":  "Chevy",
		"4":  "Dodge",
		"5":  "Ferrari",
		"6":  "Ford",
		"7":  "GM",
		"8":  "Hyundai",
		"9":  "Jeep",
		"10": "Land Rover",
	}

	ts = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var d []string
		ids := r.URL.Query()["id"]
		for _, id := range ids {
			lock.Lock()
			d = append(d, data[id])
			lock.Unlock()
		}
		b, _ := json.MarshalIndent(d, "", "  ")
		fmt.Fprintln(w, string(b))
	}))
}

// the only thing we need is a batch function that follow this signature
func batchUsers(keys []string) (results []*Result) {
	var brands []string
	v := url.Values{}
	for _, key := range keys {
		v.Add("id", key)
	}
	queryString := v.Encode()
	res, err := http.Get(ts.URL + "?" + queryString)
	if err != nil {
		// do something
	}
	if err := json.NewDecoder(res.Body).Decode(&brands); err != nil {
		// do something
	}
	for _, brand := range brands {
		results = append(results, &Result{brand, nil})
	}
	return
}

func TestLoader(t *testing.T) {
	cache := NewCache()
	UserLoader := NewBatchedLoader(batchUsers, time.Duration(16*time.Millisecond), cache, 0)

	UserLoader.Prime("cachedId", "TEST BRAND")

	future1 := UserLoader.Load("1")
	future2 := UserLoader.Load("2")
	future3 := UserLoader.Load("cachedId")

	value1 := <-future1
	value2 := <-future2
	value3 := <-future3

	// should be cached by this point
	future4 := UserLoader.Load("1")
	value4 := <-future4

	listFuture := UserLoader.LoadMany([]string{"3", "4"})
	values := <-listFuture

	log.Printf("test1: %#v", value1)
	log.Printf("test2: %#v", value2)
	log.Printf("test3: %#v", value3)
	log.Printf("test4: %#v", value4)
	log.Printf("test many: %#v", values)
}
