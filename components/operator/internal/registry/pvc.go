package registry

import (
	"context"
	"fmt"

	"github.com/pkg/errors"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const pvcName = "serverless-docker-registry"

func GetClaimedServerlessDockerRegistryStorageSize(ctx context.Context, k8sClient client.Client, namespace string) (*resource.Quantity, error) {
	pvc, err := getPVC(ctx, k8sClient, namespace, pvcName)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("while getting %s pvc", pvcName))
	}
	if pvc == nil {
		return nil, nil
	}
	size := pvc.Spec.Resources.Requests["storage"]
	return &size, nil
}

func GetRealServerlessDockerRegistryStorageSize(ctx context.Context, k8sClient client.Client, namespace string) (*resource.Quantity, error) {
	pvc, err := getPVC(ctx, k8sClient, namespace, pvcName)
	if err != nil {
		return nil, errors.Wrap(err, fmt.Sprintf("while getting %s pvc", pvcName))
	}
	if pvc == nil {
		return nil, nil
	}
	return pvc.Status.Capacity.Storage(), nil
}

func getPVC(ctx context.Context, k8sClient client.Client, namespace, name string) (*corev1.PersistentVolumeClaim, error) {
	pvc := corev1.PersistentVolumeClaim{}
	err := k8sClient.Get(ctx, client.ObjectKey{Namespace: namespace, Name: name}, &pvc)
	if err != nil {
		if apierrors.IsNotFound(err) {
			return nil, nil
		}
		return nil, errors.Wrap(err, fmt.Sprintf("while getting %s pvc", name))
	}
	return &pvc, nil
}
