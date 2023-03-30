package chart

import (
	"fmt"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

func CheckCRDOrphanResources(config *Config) error {
	manifest, err := getManifest(config)
	if err != nil {
		return fmt.Errorf("could not render manifest from chart: %s", err.Error())
	}

	objs, err := parseManifest(manifest)
	if err != nil {
		return fmt.Errorf("could not parse chart manifest: %s", err.Error())
	}

	for _, obj := range objs {
		if !isCRD(obj) {
			continue
		}

		crList, err := buildResourceListFromCRD(obj)
		if err != nil {
			return err
		}

		err = config.Cluster.Client.List(config.Ctx, &crList)
		if err != nil {
			return err
		}

		if len(crList.Items) > 0 {
			return fmt.Errorf("found %d items with VersionKind %s", len(crList.Items), crList.GetAPIVersion())
		}
	}

	return nil
}

func isCRD(u unstructured.Unstructured) bool {
	return u.GroupVersionKind().GroupKind() == apiextensionsv1.Kind("CustomResourceDefinition")
}

func buildResourceListFromCRD(u unstructured.Unstructured) (unstructured.UnstructuredList, error) {
	crd := apiextensionsv1.CustomResourceDefinition{}
	crdList := unstructured.UnstructuredList{}

	err := runtime.DefaultUnstructuredConverter.FromUnstructured(u.Object, &crd)
	if err != nil {
		return crdList, err
	}

	crdList.SetGroupVersionKind(schema.GroupVersionKind{
		Group:   crd.Spec.Group,
		Version: getCRDStoredVersion(crd),
		Kind:    crd.Spec.Names.Kind,
	})

	return crdList, nil
}

func getCRDStoredVersion(crd apiextensionsv1.CustomResourceDefinition) string {
	for _, version := range crd.Spec.Versions {
		if version.Storage {
			return version.Name
		}
	}

	return ""
}
