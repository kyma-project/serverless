package chart

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type FilterFunc func(unstructured.Unstructured) bool

func Uninstall(config *Config, filterFunc ...FilterFunc) error {
	spec, err := config.Cache.Get(config.Ctx, config.CacheKey)
	if err != nil {
		return fmt.Errorf("could not render manifest from chart: %s", err.Error())
	}

	objs, err := parseManifest(spec.Manifest)
	if err != nil {
		return fmt.Errorf("could not parse chart manifest: %s", err.Error())
	}

	for i := range objs {
		u := objs[i]
		if !fitToFilters(u, filterFunc...) {
			continue
		}

		config.Log.Debugf("deleting %s %s", u.GetKind(), u.GetName())
		err := config.Cluster.Client.Delete(config.Ctx, &u)
		if errors.IsNotFound(err) {
			config.Log.Debugf("deletion skipped for %s %s", u.GetKind(), u.GetName())
			continue
		}
		if err != nil {
			return fmt.Errorf("could not uninstall object %s/%s: %s", u.GetNamespace(), u.GetName(), err.Error())
		}
	}

	return config.Cache.Delete(config.Ctx, config.CacheKey)
}

func WithoutCRDFilter(u unstructured.Unstructured) bool {
	return !isCRD(u)
}

func fitToFilters(u unstructured.Unstructured, filterFunc ...FilterFunc) bool {
	for _, fn := range filterFunc {
		if !fn(u) {
			return false
		}
	}

	return true
}
