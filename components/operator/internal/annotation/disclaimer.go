package annotation

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/cli-runtime/pkg/resource"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	reconcilerPatch = "{\"metadata\":{\"annotations\":{\"reconciler.kyma-project.io/managed-by-reconciler-disclaimer\":null}}}"
	annotation      = "serverless-manager.kyma-project.io/managed-by-serverless-manager-disclaimer"
	message         = "DO NOT EDIT - This resource is managed by Serverless-Manager.\nAny modifications are discarded and the resource is reverted to the original state."
)

func AddDoNotEditDisclaimer(obj unstructured.Unstructured) unstructured.Unstructured {
	annotations := obj.GetAnnotations()
	if annotations == nil {
		annotations = map[string]string{}
	}

	annotations[annotation] = message
	obj.SetAnnotations(annotations)

	return obj
}

func DeleteReconcilerDisclaimer(client client.Client, config rest.Config, obj unstructured.Unstructured) error {
	mapping, err := client.RESTMapper().RESTMapping(
		obj.GroupVersionKind().GroupKind(),
		obj.GroupVersionKind().Version,
	)
	if err != nil {
		return err
	}

	restClient, err := UnstructuredClientForMapping(&config, mapping)
	if err != nil {
		return err
	}

	helper := resource.
		NewHelper(restClient, mapping).
		DryRun(false).
		WithFieldManager("serverless-manager")

	_, err = helper.Patch(obj.GetNamespace(), obj.GetName(), types.MergePatchType, []byte(reconcilerPatch), nil)
	return err
}

func UnstructuredClientForMapping(config *rest.Config, mapping *metav1.RESTMapping) (resource.RESTClient, error) {
	config.APIPath = "/apis"
	if mapping.GroupVersionKind.Group == corev1.GroupName {
		config.APIPath = "/api"
	}
	gv := mapping.GroupVersionKind.GroupVersion()
	config.ContentConfig = resource.UnstructuredPlusDefaultContentConfig()
	config.GroupVersion = &gv
	return rest.RESTClientFor(config)
}
