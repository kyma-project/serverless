package state

import (
	"testing"

	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func Test_sFnRemoveFinalizer(t *testing.T) {
	t.Run("remove finalizer", func(t *testing.T) {
		r := &reconciler{
			cfg: cfg{
				finalizer: v1alpha1.Finalizer,
			},
		}
		s := &systemState{
			instance: v1alpha1.Serverless{
				ObjectMeta: metav1.ObjectMeta{
					Finalizers: []string{
						r.finalizer,
					},
				},
			},
		}

		// remove finalizer
		next, result, err := sFnRemoveFinalizer()(nil, r, s)

		expectedNext := sFnUpdateServerless()

		requireEqualFunc(t, expectedNext, next)
		require.Nil(t, result)
		require.Nil(t, err)
	})

	t.Run("requeue when is no finalizer", func(t *testing.T) {
		r := &reconciler{
			cfg: cfg{
				finalizer: v1alpha1.Finalizer,
			},
		}
		s := &systemState{
			instance: v1alpha1.Serverless{
				ObjectMeta: metav1.ObjectMeta{},
			},
		}

		// remove finalizer
		next, result, err := sFnRemoveFinalizer()(nil, r, s)

		expectedNext := sFnEmitStrictEvent(
			nil, nil, nil,
			"Normal",
			"Deleted",
			"Serverless module deleted",
		)

		requireEqualFunc(t, expectedNext, next)
		require.Nil(t, result)
		require.Nil(t, err)
	})
}
