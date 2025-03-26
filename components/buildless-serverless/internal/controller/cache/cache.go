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

// inMemoryManifestCache provides an in-memory processor to store serverless Spec and rendered chart manifest. By using sync.Map for caching,
// concurrent operations to the processor from diverse reconciliations are considered safe.
//
// Inside the processor is stored chart manifest with used custom flags by client.ObjectKey key.
type inMemoryCache struct {
	processor sync.Map
	timeout   time.Duration
}

type storageObject struct {
	timestamp time.Time
	value     string
}

// NewInMemoryManifestCache returns a new instance of inMemoryManifestCache.
func NewInMemoryCache(timeout time.Duration) *inMemoryCache {
	return &inMemoryCache{
		processor: sync.Map{},
		timeout:   timeout,
	}
}

// Get loads the ServerlessSpecManifest from inMemoryManifestCache for the passed client.ObjectKey.
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

// Set saves the passed flags and manifest into inMemoryManifestCache for the client.ObjectKey.
func (r *inMemoryCache) Set(key any, lastCommit string) {
	now := time.Now()
	r.processor.Store(key, &storageObject{now, lastCommit})
}

// Delete deletes flags and manifest from inMemoryManifestCache for the passed client.ObjectKey.
func (r *inMemoryCache) Delete(key any) {
	r.processor.Delete(key)
}
