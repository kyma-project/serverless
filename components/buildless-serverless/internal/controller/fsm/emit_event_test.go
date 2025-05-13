package fsm

import (
	"testing"

	"github.com/kyma-project/serverless/components/buildless-serverless/api/v1alpha2"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
)

var (
	testFunctionConditions1 = v1alpha2.Function{
		Status: v1alpha2.FunctionStatus{
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
	testFunctionConditions2 = v1alpha2.Function{
		Status: v1alpha2.FunctionStatus{
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
		sm := &StateMachine{
			State: SystemState{
				Function:       *testFunctionConditions1.DeepCopy(),
				statusSnapshot: *testFunctionConditions1.Status.DeepCopy(),
			},
			EventRecorder: eventRecorder,
		}

		emitEvent(sm)

		require.Len(t, eventRecorder.Events, 0)
	})

	t.Run("emit events", func(t *testing.T) {
		eventRecorder := record.NewFakeRecorder(5)
		sm := &StateMachine{
			State: SystemState{
				Function:       *testFunctionConditions2.DeepCopy(),
				statusSnapshot: *testFunctionConditions1.Status.DeepCopy(),
			},
			EventRecorder: eventRecorder,
		}

		emitEvent(sm)

		require.Len(t, eventRecorder.Events, 2)
		expectedEvents := []string{"Warning test-reason test message 2", "Normal test-reason test message 2"}
		close(eventRecorder.Events)
		for v := range eventRecorder.Events {
			require.Contains(t, expectedEvents, v)
		}
	})
}
