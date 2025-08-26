package serverless

import (
	"context"
	"fmt"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func DeleteIstioNativeSidecar(ctx context.Context, m manager.Manager) error {
	m.GetLogger().Info("Deleting Istio native sidecar annotations from Functions")

	annotation := "sidecar.istio.io/nativeSidecar"

	var collectedErrors []string

	// list pods with the specific annotation
	pods := &corev1.PodList{}
	err := listAnnotated(ctx, m.GetAPIReader(), annotation, pods)
	if err != nil {
		collectedErrors = append(collectedErrors, fmt.Sprintf("failed to list annotated pods: %s", err))
	}

	// delete the annotation from each pod
	for i := range pods.Items {
		pod := &pods.Items[i]
		base := pod.DeepCopy()

		if pod.Annotations != nil {
			delete(pod.Annotations, annotation)
		}

		if err := m.GetClient().Patch(ctx, pod, client.MergeFrom(base)); err != nil {
			collectedErrors = append(collectedErrors,
				fmt.Sprintf("failed to delete annotation from pod %s/%s: %s",
					pod.Namespace, pod.Name, err))
		}
	}

	// list deployments with the specific annotation
	deployments := &appsv1.DeploymentList{}
	err = listAnnotated(ctx, m.GetAPIReader(), annotation, deployments)
	if err != nil {
		collectedErrors = append(collectedErrors, fmt.Sprintf("failed to list annotated deployments: %s", err))
	}

	// delete the annotation from each deployment
	for i := range deployments.Items {
		deployment := &deployments.Items[i]
		base := deployment.DeepCopy()
		// Remove annotation from Deployment metadata
		if deployment.Annotations != nil {
			m.GetLogger().Info("Removing annotation from deployment",
				"namespace", deployment.Namespace, "name", deployment.Name)
			delete(deployment.Annotations, annotation)
		}
		// Remove annotation from Deployment pod template
		if deployment.Spec.Template.ObjectMeta.Annotations != nil {
			m.GetLogger().Info("Removing annotation from deployment",
				"namespace", deployment.Namespace, "name", deployment.Name)
			delete(deployment.Spec.Template.ObjectMeta.Annotations, annotation)
		}
		if err := m.GetClient().Patch(ctx, deployment, client.MergeFrom(base)); err != nil {
			collectedErrors = append(collectedErrors, fmt.Sprintf("failed to delete annotation from deployment %s/%s: %s", deployment.Namespace, deployment.Name, err))
		}
	}

	if len(collectedErrors) > 0 {
		return fmt.Errorf("errors occurred while deleting Istio native sidecar annotations: %v", collectedErrors)
	}

	return nil
}

func listAnnotated(ctx context.Context, reader client.Reader, annotation string, list client.ObjectList) error {
	if err := reader.List(ctx, list, &client.ListOptions{}); err != nil {
		return err
	}

	// Filter objects with the specific annotation
	items, err := meta.ExtractList(list)
	if err != nil {
		return err
	}

	filteredItems := []runtime.Object{}
	for _, item := range items {
		obj := item.(client.Object)
		if _, exists := obj.GetAnnotations()[annotation]; exists {
			filteredItems = append(filteredItems, obj)
		}
	}

	return meta.SetList(list, filteredItems)
}
