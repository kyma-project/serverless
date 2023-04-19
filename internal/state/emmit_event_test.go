package state

import (
	"testing"

	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	"github.com/stretchr/testify/require"
	"k8s.io/client-go/tools/record"
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
