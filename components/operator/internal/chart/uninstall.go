package chart

import (
	"fmt"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
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

	err2 := uninstallObjects(config, objs, filterFunc...)
	if err2 != nil {
		return err2
	}

	err3 := uninstallOrphanedResources(config)
	if err3 != nil {
		return err3
	}

	return config.Cache.Delete(config.Ctx, config.CacheKey)
}

func uninstallObjects(config *Config, objs []unstructured.Unstructured, filterFunc ...FilterFunc) error {
	for i := range objs {
		u := objs[i]
		if !fitToFilters(u, filterFunc...) {
			continue
		}

		config.Log.Debugf("deleting %s %s", u.GetKind(), u.GetName())
		err := config.Cluster.Client.Delete(config.Ctx, &u)
		if k8serrors.IsNotFound(err) {
			config.Log.Debugf("deletion skipped for %s %s", u.GetKind(), u.GetName())
			continue
		}
		if err != nil {
			return fmt.Errorf("could not uninstall object %s/%s: %s", u.GetNamespace(), u.GetName(), err.Error())
		}
	}
	return nil
}

func UninstallResourcesByType(config *Config, resourceType string, filterFunc ...FilterFunc) (error, bool) {
	spec, err := config.Cache.Get(config.Ctx, config.CacheKey)
	if err != nil {
		return fmt.Errorf("could not render manifest from chart: %s", err.Error()), false
	}

	objs, err := parseManifest(spec.Manifest)
	if err != nil {
		return fmt.Errorf("could not parse chart manifest: %s", err.Error()), false
	}

	err2, done := uninstallResourcesByType(config, objs, resourceType, filterFunc...)
	if err2 != nil {
		return err2, false
	}

	return nil, done
}

func uninstallResourcesByType(config *Config, objs []unstructured.Unstructured, resourceType string, filterFunc ...FilterFunc) (error, bool) {
	done := true
	for i := range objs {
		u := objs[i]
		if !fitToFilters(u, filterFunc...) {
			continue
		}
		if u.GetKind() != resourceType {
			continue
		}

		config.Log.Debugf("deleting %s %s", u.GetKind(), u.GetName())
		err := config.Cluster.Client.Delete(config.Ctx, &u)
		if k8serrors.IsNotFound(err) {
			config.Log.Debugf("deletion skipped for %s %s", u.GetKind(), u.GetName())
			continue
		}
		if err != nil {
			return fmt.Errorf("could not uninstall object %s/%s: %s", u.GetNamespace(), u.GetName(), err.Error()), false
		}
		done = false
	}
	return nil, done
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

func uninstallOrphanedResources(config *Config) error {
	//TODO: move this to finalizers logic in controller
	var namespaces corev1.NamespaceList
	if err := config.Cluster.Client.List(config.Ctx, &namespaces); err != nil {
		return errors.Wrap(err, "couldn't get namespaces during Serverless uninstallation")
	}

	if err := uninstallOrphanedConfigMaps(config, namespaces); err != nil {
		return err
	}
	if err := uninstallOrphanedServiceAccounts(config, namespaces); err != nil {
		return err
	}

	return nil
}

func uninstallOrphanedServiceAccounts(config *Config, namespaces corev1.NamespaceList) error {
	for _, namespace := range namespaces.Items {
		err := config.Cluster.Client.DeleteAllOf(config.Ctx, &corev1.ServiceAccount{},
			client.InNamespace(namespace.GetName()),
			client.MatchingLabels{"serverless.kyma-project.io/config": "service-account"})
		if err != nil {
			return errors.Wrapf(err,
				"couldn't delete ServiceAccount from namespace \"%s\" during Serverless uninstallation",
				namespace.GetName())
		}
	}
	return nil
}

func uninstallOrphanedConfigMaps(config *Config, namespaces corev1.NamespaceList) error {
	for _, namespace := range namespaces.Items {
		err := config.Cluster.Client.DeleteAllOf(config.Ctx, &corev1.ConfigMap{},
			client.InNamespace(namespace.GetName()),
			client.MatchingLabels{"serverless.kyma-project.io/config": "runtime"})
		if err != nil {
			return errors.Wrapf(err,
				"couldn't delete ConfigMap from namespace \"%s\" during Serverless uninstallation",
				namespace.GetName())
		}
	}
	return nil
}
