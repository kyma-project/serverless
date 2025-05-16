package job

import (
	"context"
	"fmt"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apilabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/manager"
)

func DeleteOrphanedJobs(ctx context.Context, m manager.Manager) error {
	labels := map[string]string{
		"serverless.kyma-project.io/managed-by": "function-controller",
	}
	// list orphaned jobs
	jobs := &batchv1.JobList{}
	err := m.GetAPIReader().List(ctx, jobs, &client.ListOptions{
		LabelSelector: apilabels.SelectorFromSet(labels),
	})
	if err != nil {
		return fmt.Errorf("failed to list orphaned jobs: %w", err)
	}

	for _, job := range jobs.Items {
		err := m.GetClient().Delete(ctx, &job, &client.DeleteOptions{
			PropagationPolicy: ptr.To(metav1.DeletePropagationBackground),
		})
		if err != nil {
			return err
		}
	}
	return err
}
