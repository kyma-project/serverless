package cache

import (
	"sync"
	"time"
)

var (
	_ InMemoryCache = (*inMemoryCache)(nil)
)

type InMemoryCache interface {
	Set(any, string)
	Get(any) string
	Delete(any)
}

// inMemoryCache provides an in-memory processor to store buildless serverless key values pairs with timestamp. By using sync.Map for caching,
// concurrent operations to the processor from diverse reconciliations are considered safe.
//
// Inside the processor is stored last commit with which was git function created by repository url and reference as the key.
type inMemoryCache struct {
	processor sync.Map
	timeout   time.Duration
}

type storageObject struct {
	timestamp time.Time
	value     string
}

// NewInMemoryCache returns a new instance of inMemoryCache.
func NewInMemoryCache(timeout time.Duration) *inMemoryCache {
	return &inMemoryCache{
		processor: sync.Map{},
		timeout:   timeout,
	}
}

// Get loads from inMemoryCache for the passed key.
func (r *inMemoryCache) Get(key any) string {
	rawValue, ok := r.processor.Load(key)
	if !ok {
		return ""
	}
	value := *rawValue.(*storageObject)
	if time.Since(value.timestamp) > r.timeout {
		return ""
	}

	return value.value
}

// Set saves the passed last commit with which was git function created in inMemoryCache for the passed key.
func (r *inMemoryCache) Set(key any, lastCommit string) {
	now := time.Now()
	r.processor.Store(key, &storageObject{now, lastCommit})
}

// Delete deletes last commit with which was git function created from inMemoryCache for the passed key.
func (r *inMemoryCache) Delete(key any) {
	r.processor.Delete(key)
}
