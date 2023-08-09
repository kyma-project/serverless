package chart

import (
	"context"
	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func AdjustToClusterSize(ctx context.Context, c client.Client, obj unstructured.Unstructured) (unstructured.Unstructured, error) {
	clusterPVC := corev1.PersistentVolumeClaim{}

	err := c.Get(ctx, client.ObjectKey{
		Namespace: obj.GetNamespace(),
		Name:      obj.GetName(),
	}, &clusterPVC)
	if err != nil {
		if k8serrors.IsNotFound(err) {
			return obj, nil
		}
	}
	objPVC := corev1.PersistentVolumeClaim{}
	if err := runtime.DefaultUnstructuredConverter.FromUnstructured(obj.Object, &objPVC); err != nil {
		return obj, errors.Wrap(err, "while converting unstructured to pvc")
	}
	storage := clusterPVC.Spec.Resources.Requests.Storage()
	if storage.Equal(*objPVC.Spec.Resources.Requests.Storage()) {
		return obj, nil
	}
	objPVCcopy := objPVC.DeepCopy()

	objPVCcopy.Spec.Resources.Requests[corev1.ResourceStorage] = *clusterPVC.Spec.Resources.Requests.Storage()

	out, err := runtime.DefaultUnstructuredConverter.ToUnstructured(objPVCcopy)
	if err != nil {
		return obj, errors.Wrap(err, "while converting copied pvc object to unstructured")
	}

	return unstructured.Unstructured{Object: out}, nil
}

func IsPVC(objKind schema.GroupVersionKind) bool {
	claim := corev1.PersistentVolumeClaim{}
	expected := claim.GroupVersionKind()

	return expected.Group == objKind.Group && expected.Kind == objKind.Kind && expected.Version == objKind.Version
}
