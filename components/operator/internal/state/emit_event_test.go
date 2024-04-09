package state

import (
	"testing"

	"github.com/kyma-project/serverless/components/operator/api/v1alpha1"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
)

var (
	testDockerRegistryConditions1 = v1alpha1.DockerRegistry{
		Status: v1alpha1.DockerRegistryStatus{
			Conditions: []metav1.Condition{
				{
					Status:  metav1.ConditionUnknown,
					Reason:  "test-reason",
					Message: "test message 1",
					Type:    "test-type-1",
				},
				{
					Status:  metav1.ConditionUnknown,
					Reason:  "test-reason",
					Message: "test message 1",
					Type:    "test-type-2",
				},
			},
		},
	}
	testDockerRegistryConditions2 = v1alpha1.DockerRegistry{
		Status: v1alpha1.DockerRegistryStatus{
			Conditions: []metav1.Condition{
				{
					Status:  metav1.ConditionFalse,
					Reason:  "test-reason",
					Message: "test message 2",
					Type:    "test-type-1",
				},
				{
					Status:  metav1.ConditionTrue,
					Reason:  "test-reason",
					Message: "test message 2",
					Type:    "test-type-2",
				},
			},
		},
	}
)

func Test_emitEvent(t *testing.T) {
	t.Run("don't emit event", func(t *testing.T) {
		eventRecorder := record.NewFakeRecorder(5)
		s := &systemState{
			instance:       *testDockerRegistryConditions1.DeepCopy(),
			statusSnapshot: *testDockerRegistryConditions1.Status.DeepCopy(),
		}
		r := &reconciler{
			k8s: k8s{
				EventRecorder: eventRecorder,
			},
		}

		emitEvent(r, s)

		// check conditions, don't emit event
		require.Len(t, eventRecorder.Events, 0)
	})

	t.Run("emit events", func(t *testing.T) {
		eventRecorder := record.NewFakeRecorder(5)
		s := &systemState{
			instance:       *testDockerRegistryConditions2.DeepCopy(),
			statusSnapshot: *testDockerRegistryConditions1.Status.DeepCopy(),
		}
		r := &reconciler{
			k8s: k8s{
				EventRecorder: eventRecorder,
			},
		}

		// build emitEventFunc
		emitEvent(r, s)

		// check conditions, don't emit event
		require.Len(t, eventRecorder.Events, 2)

		expectedEvents := []string{"Warning test-reason test message 2", "Normal test-reason test message 2"}
		close(eventRecorder.Events)
		for v := range eventRecorder.Events {
			require.Contains(t, expectedEvents, v)
		}
	})
}
