package orphaned_resources

import (
	"context"
	"fmt"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apilabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func DeleteOrphanedResources(ctx context.Context, m manager.Manager) error {
	labels := map[string]string{
		"serverless.kyma-project.io/managed-by": "function-controller",
	}

	// list orphaned jobs
	jobs := &batchv1.JobList{}
	err := listOrphanedResources(ctx, m.GetAPIReader(), jobs, labels)
	if err != nil {
		return fmt.Errorf("failed to list orphaned jobs: %s", err)
	}

	// delete orphaned jobs
	for _, job := range jobs.Items {
		err := deleteOrphanedResource(ctx, m.GetClient(), &job)
		if err != nil {
			return fmt.Errorf("failed to delete orphaned-resources %s/%s: %s", job.Namespace, job.Name, err)
		}
	}

	// list orphaned configmaps
	configMaps := &corev1.ConfigMapList{}
	err = listOrphanedResources(ctx, m.GetAPIReader(), configMaps, labels)
	if err != nil {
		return fmt.Errorf("failed to list orphaned configmaps: %s", err)
	}

	// delete orphaned configmaps
	for _, configMap := range configMaps.Items {
		err := deleteOrphanedResource(ctx, m.GetClient(), &configMap)
		if err != nil {
			return fmt.Errorf("failed to delete orphaned configmap %s/%s: %s", configMap.Namespace, configMap.Name, err)
		}
	}

	return nil
}

func listOrphanedResources(ctx context.Context, m client.Reader, resourceList client.ObjectList, labels map[string]string) error {
	return m.List(ctx, resourceList, &client.ListOptions{
		LabelSelector: apilabels.SelectorFromSet(labels),
	})
}

func deleteOrphanedResource(ctx context.Context, m client.Client, resource client.Object) error {
	return m.Delete(ctx, resource, &client.DeleteOptions{
		PropagationPolicy: ptr.To(metav1.DeletePropagationBackground),
	})
}
