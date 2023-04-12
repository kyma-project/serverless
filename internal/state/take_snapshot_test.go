package state

import (
	"testing"

	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_sFnTakeSnapshot(t *testing.T) {
	t.Run("take snapshot", func(t *testing.T) {
		serverless := v1alpha1.Serverless{
			Status: v1alpha1.ServerlessStatus{
				Conditions: []metav1.Condition{
					{
						Type:               "test-type",
						Status:             "test-status",
						Reason:             "test-reason",
						Message:            "test-message",
						ObservedGeneration: 1,
						LastTransitionTime: metav1.Now(),
					},
				},
				State: v1alpha1.StateError,
			},
		}
		s := &systemState{
			instance: serverless,
		}

		// build sFn
		sFn := buildSFnTakeSnapshot(sFnInitialize, nil, nil)
		require.NotNil(t, sFn)

		// run sFn and return sFnInitialize
		next, result, err := sFn(nil, nil, s)
		requireEqualFunc(t, sFnInitialize, next)
		require.Nil(t, result)
		require.NoError(t, err)

		// check status
		require.Equal(t, serverless.Status, s.snapshot)
	})

	t.Run("empty status", func(t *testing.T) {
		serverless := v1alpha1.Serverless{}
		s := &systemState{
			instance: serverless,
		}

		// build sFn
		sFn := buildSFnTakeSnapshot(nil, nil, nil)
		require.NotNil(t, sFn)

		// run sFn and return nil
		next, result, err := sFn(nil, nil, s)
		require.Nil(t, next)
		require.Nil(t, result)
		require.NoError(t, err)

		require.Equal(t, v1alpha1.ServerlessStatus{}, s.snapshot)
	})
}
