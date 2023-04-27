package state

import (
	"testing"

	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/record"
)

var (
	testServerlessConditions1 = v1alpha1.Serverless{
		Status: v1alpha1.ServerlessStatus{
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
	testServerlessConditions2 = v1alpha1.Serverless{
		Status: v1alpha1.ServerlessStatus{
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

func Test_sFnEmitStrictEvent(t *testing.T) {
	t.Run("emit event", func(t *testing.T) {
		eventRecorder := record.NewFakeRecorder(1)
		r := &reconciler{
			k8s: k8s{
				EventRecorder: eventRecorder,
			},
		}

		s := &systemState{
			instance: v1alpha1.Serverless{},
		}

		stateFn := sFnEmitStrictEvent(
			nil, nil, nil,
			"test-type",
			"test-reason",
			"test-message",
		)

		next, result, err := stateFn(nil, r, s)
		require.Nil(t, next)
		require.Nil(t, result)
		require.Nil(t, err)

		require.Len(t, eventRecorder.Events, 1)
	})
}

func Test_sFnEmitEventfunc(t *testing.T) {
	t.Run("don't emit event", func(t *testing.T) {
		s := &systemState{
			instance: testServerlessConditions1,
			snapshot: *testServerlessConditions1.Status.DeepCopy(),
		}

		// build emitEventFunc
		stateFn := buildSFnEmitEvent(nil, nil, nil)

		// check conditions, don't emit event
		next, result, err := stateFn(nil, nil, s)

		expectedNext := buildSFnTakeSnapshot(nil, nil, nil)
		requireEqualFunc(t, expectedNext, next)
		require.Nil(t, result)
		require.Nil(t, err)
	})

	t.Run("emit events", func(t *testing.T) {
		eventRecorder := record.NewFakeRecorder(2)

		s := &systemState{
			instance: testServerlessConditions2,
			snapshot: *testServerlessConditions1.Status.DeepCopy(),
		}

		r := &reconciler{
			k8s: k8s{
				EventRecorder: eventRecorder,
			},
		}

		// build emitEventFunc
		stateFn := buildSFnEmitEvent(nil, nil, nil)

		// check conditions, don't emit event
		next, result, err := stateFn(nil, r, s)

		expectedNext := buildSFnTakeSnapshot(nil, nil, nil)
		requireEqualFunc(t, expectedNext, next)
		require.Nil(t, result)
		require.Nil(t, err)

		require.Len(t, eventRecorder.Events, 2)
	})
}
