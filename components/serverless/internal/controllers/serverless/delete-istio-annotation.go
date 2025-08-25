package serverless

import (
	"context"
	"fmt"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func DeleteIstioNativeSidecar(ctx context.Context, m manager.Manager) error {
	m.GetLogger().Info("Deleting Istio native sidecar annotations from Functions")

	annotation := "sidecar.istio.io/nativeSidecar"

	var collectedErrors []string

	// list pods with the specific annotation
	pods, err := listAnnotatedPods(ctx, m.GetAPIReader(), annotation)
	if err != nil {
		collectedErrors = append(collectedErrors, fmt.Sprintf("failed to list annotated pods: %s", err))
	}

	// delete the annotation from each pod
	for _, pod := range pods.Items {
		patch := client.MergeFrom(pod.DeepCopy())
		delete(pod.Annotations, annotation)
		err := m.GetClient().Patch(ctx, &pod, patch)
		if err != nil {
			collectedErrors = append(collectedErrors, fmt.Sprintf("failed to delete annotation from pod %s/%s: %s", pod.Namespace, pod.Name, err))
		}
	}

	if len(collectedErrors) > 0 {
		return fmt.Errorf("errors occurred while deleting Istio native sidecar annotations: %v", collectedErrors)
	}

	return nil
}

func listAnnotatedPods(ctx context.Context, m client.Reader, annotation string) (*corev1.PodList, error) {
	pods := &corev1.PodList{}
	err := m.List(ctx, pods, &client.ListOptions{})
	if err != nil {
		return nil, err
	}

	// Filter pods that have the specific annotation
	filteredPods := &corev1.PodList{}
	for _, pod := range pods.Items {
		if _, exists := pod.Annotations[annotation]; exists {
			filteredPods.Items = append(filteredPods.Items, pod)
		}
	}

	return filteredPods, nil
}
