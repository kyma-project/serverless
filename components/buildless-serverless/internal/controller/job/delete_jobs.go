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
	err := listJobs(ctx, m.GetAPIReader(), jobs, labels)
	if err != nil {
		return fmt.Errorf("failed to list orphaned jobs: %s", err)
	}

	// delete orphaned jobs
	for _, job := range jobs.Items {
		err := deleteJobs(ctx, m.GetClient(), &job)
		if err != nil {
			return fmt.Errorf("failed to delete job %s/%s: %s", job.Namespace, job.Name, err)
		}
	}

	return nil
}

func listJobs(ctx context.Context, m client.Reader, jobs *batchv1.JobList, labels map[string]string) error {
	return m.List(ctx, jobs, &client.ListOptions{
		LabelSelector: apilabels.SelectorFromSet(labels),
	})
}

func deleteJobs(ctx context.Context, m client.Client, job *batchv1.Job) error {
	return m.Delete(ctx, job, &client.DeleteOptions{
		PropagationPolicy: ptr.To(metav1.DeletePropagationBackground),
	})
}
