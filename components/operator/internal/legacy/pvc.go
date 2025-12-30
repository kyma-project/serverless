package legacy

import (
	"context"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	dockerRegistryPVCName = "serverless-docker-registry"
)

func AdjustDockerRegToClusterPVCSize(ctx context.Context, c client.Client, obj unstructured.Unstructured) (unstructured.Unstructured, error) {
	if obj.GetName() != dockerRegistryPVCName {
		return obj, nil
	}
	clusterPVC := corev1.PersistentVolumeClaim{}
	objKey := client.ObjectKey{
		Namespace: obj.GetNamespace(),
		Name:      obj.GetName(),
	}
	if err := c.Get(ctx, objKey, &clusterPVC); err != nil {
		if k8serrors.IsNotFound(err) {
			return obj, nil
		}
		return obj, errors.Wrap(err, "while getting pvc from cluster")
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
