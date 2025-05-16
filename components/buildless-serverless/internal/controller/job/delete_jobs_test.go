package job

import (
	"context"
	"github.com/stretchr/testify/require"
	batchv1 "k8s.io/api/batch/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func Test_listJobs(t *testing.T) {
	t.Run("should list jobs with matching labels", func(t *testing.T) {
		// Arrange
		correctLabels := map[string]string{
			"serverless.kyma-project.io/managed-by": "function-controller",
		}
		incorrectLabels := map[string]string{
			"serverless.kyma-project.io/not-managed-by": "function-controller",
		}


		job1 := batchv1.Job{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "test-job1",
						Namespace: metav1.NamespaceDefault,
						Labels:    correctLabels,
					},
				}
		job2 := batchv1.Job{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-job2",
					Namespace: "another-namespace",
					Labels:    correctLabels,
				},
			}
		someJob := batchv1.Job{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-job3",
				Namespace: metav1.NamespaceDefault,
				Labels:    incorrectLabels,
			},
		}



		scheme := runtime.NewScheme()
		require.NoError(t, batchv1.AddToScheme(scheme))
		k8sClient := fake.NewClientBuilder().WithScheme(scheme).WithObjects(&job1, &job2, &someJob).Build()

		// Act
		err := listJobs(ctx, m, jobs, labels)

		// Assert
		if err != nil {
			t.Errorf("listJobs() error = %v", err)
		}
		if len(jobs.Items) == 0 {
			t.Errorf("listJobs() expected to find jobs, but found none")
		}
	})
	}
}

func Test_deleteJobs(t *testing.T) {
	type args struct {
		ctx context.Context
		m   client.Client
		job *batchv1.Job
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := deleteJobs(tt.args.ctx, tt.args.m, tt.args.job); (err != nil) != tt.wantErr {
				t.Errorf("deleteJobs() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}