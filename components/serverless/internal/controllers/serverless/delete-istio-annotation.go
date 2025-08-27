package serverless

import (
	"context"
	"fmt"
	serverlessv1alpha2 "github.com/kyma-project/serverless/components/serverless/pkg/apis/serverless/v1alpha2"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func DeleteIstioNativeSidecar(ctx context.Context, m manager.Manager) error {
	m.GetLogger().Info("Deleting Istio native sidecar annotations from Functions")

	annotation := "sidecar.istio.io/nativeSidecar"

	var collectedErrors []string

	// list deployments with the specific annotation
	deployments := &appsv1.DeploymentList{}
	var f serverlessv1alpha2.Function
	err := listAnnotated(annotation, *deployments, m, systemState{instance: f})
	if err != nil {
		collectedErrors = append(collectedErrors, fmt.Sprintf("failed to list annotated deployments: %s", err))
	}

	m.GetLogger().Info(fmt.Sprintf("Length %d", len(deployments.Items)))

	// delete the annotation from each deployment
	for i := range deployments.Items {
		deployment := &deployments.Items[i]
		base := deployment.DeepCopy()
		m.GetLogger().Info("Before patch", "annotations", deployment.Spec.Template.ObjectMeta.Annotations)
		m.GetLogger().Info(fmt.Sprintf("Annotations %v, %v", deployment.Annotations, deployment.Spec.Template.ObjectMeta.Annotations))
		//base := deployment.DeepCopy()
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
		m.GetLogger().Info("After patch", "annotations", deployment.Spec.Template.ObjectMeta.Annotations)
	}

	if len(collectedErrors) > 0 {
		return fmt.Errorf("errors occurred while deleting Istio native sidecar annotations: %v", collectedErrors)
	}

	return nil
}

func listAnnotated(annotation string, list appsv1.DeploymentList, m manager.Manager, s systemState) error {
	filteredItems := []runtime.Object{}
	for _, item := range s.deployments.Items {
		if _, exists := item.Annotations[annotation]; exists {
			filteredItems = append(filteredItems, &item)
		}
	}

	m.GetLogger().Info(fmt.Sprintf("Length %d", len(filteredItems)))

	return meta.SetList(&list, filteredItems)
}
