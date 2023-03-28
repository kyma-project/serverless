package chart

import (
	"sync"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

// ManifestCache provides an in-memory processor to store serverless Spec and rendered chart manifest. By using sync.Map for caching,
// concurrent operations to the processor from diverse reconciliations are considered safe.
//
// Inside the processor is stored chart manifest with used custom flags by client.ObjectKey key.
type ManifestCache struct {
	processor sync.Map
}

type ServerlessSpecManifest struct {
	customFlags map[string]interface{}
	manifest    string
}

// NewManifestCache returns a new instance of ManifestCache.
func NewManifestCache() *ManifestCache {
	return &ManifestCache{
		processor: sync.Map{},
	}
}

// Get loads the ServerlessSpecManifest from ManifestCache for the passed client.ObjectKey.
func (r *ManifestCache) Get(key client.ObjectKey) *ServerlessSpecManifest {
	value, ok := r.processor.Load(key)
	if !ok {
		return nil
	}

	return value.(*ServerlessSpecManifest)
}

// SetProcessor saves the passed flags and manifest into ManifestCache for the client.ObjectKey.
func (r *ManifestCache) Set(key client.ObjectKey, customFlags map[string]interface{}, manifest string) {
	r.processor.Store(key, &ServerlessSpecManifest{
		customFlags: customFlags,
		manifest:    manifest,
	})
}

// DeleteProcessor deletes flags and manifest from ManifestCache for the passed client.ObjectKey.
func (r *ManifestCache) Delete(key client.ObjectKey) {
	r.processor.Delete(key)
}
