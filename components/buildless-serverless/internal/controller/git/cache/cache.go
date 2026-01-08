package cache

import (
	"sync"
	"time"
)

var (
	_ Cache = (*repoLastCommitCache)(nil)
)

type Cache interface {
	Set(any, string)
	Get(any) *string
	Delete(any)
}

// repoLastCommitCache provides an in-memory processor to store buildless serverless key values pairs with timestamp. By using sync.Map for caching,
// concurrent operations to the processor from diverse reconciliations are considered safe.
//
// Inside the processor is stored last commit with which was git function created by repository url and reference as the key.
type repoLastCommitCache struct {
	processor sync.Map
	timeout   time.Duration
}

type storageObject struct {
	timestamp time.Time
	value     string
}

// NewRepoLastCommitCache returns a new instance of repoLastCommitCache.
func NewRepoLastCommitCache(timeout time.Duration) Cache {
	return &repoLastCommitCache{
		processor: sync.Map{},
		timeout:   timeout,
	}
}

// Get loads from repoLastCommitCache for the passed key.
func (r *repoLastCommitCache) Get(key any) *string {
	rawValue, ok := r.processor.Load(key)
	if !ok {
		return nil
	}
	value := *rawValue.(*storageObject)
	if time.Since(value.timestamp) > r.timeout {
		return nil
	}

	return &value.value
}

// Set saves the passed last commit with which was git function created in repoLastCommitCache for the passed key.
func (r *repoLastCommitCache) Set(key any, lastCommit string) {
	now := time.Now()
	r.processor.Store(key, &storageObject{now, lastCommit})
}

// Delete deletes last commit with which was git function created from repoLastCommitCache for the passed key.
func (r *repoLastCommitCache) Delete(key any) {
	r.processor.Delete(key)
}
