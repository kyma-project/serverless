package chart

import (
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
)

type FilterFunc func(unstructured.Unstructured) bool

func Uninstall(config *Config, filterFunc ...FilterFunc) error {
	manifest, err := getManifest(config)
	if err != nil {
		return fmt.Errorf("could not render manifest from chart: %s", err.Error())
	}

	objs, err := parseManifest(manifest)
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

	config.Cache.Delete(types.NamespacedName{
		Name:      config.Release.Name,
		Namespace: config.Release.Namespace,
	})
	return nil
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
