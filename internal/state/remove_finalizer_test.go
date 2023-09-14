package state

import (
	"context"
	"testing"

	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func Test_sFnRemoveFinalizer(t *testing.T) {
	t.Run("remove finalizer", func(t *testing.T) {
		scheme := scheme.Scheme
		require.NoError(t, v1alpha1.AddToScheme(scheme))

		instance := v1alpha1.Serverless{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "default",
				Finalizers: []string{
					v1alpha1.Finalizer,
				},
			},
		}

		r := &reconciler{
			cfg: cfg{
				finalizer: v1alpha1.Finalizer,
			},
			k8s: k8s{
				client: fake.NewClientBuilder().
					WithScheme(scheme).
					WithObjects(&instance).
					Build(),
			},
		}
		s := &systemState{
			instance: instance,
		}

		// remove finalizer
		next, result, err := sFnRemoveFinalizer(context.Background(), r, s)

		require.Nil(t, next)
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
		next, result, err := sFnRemoveFinalizer(nil, r, s)

		require.Nil(t, next)
		require.Equal(t, &ctrl.Result{Requeue: true}, result)
		require.Nil(t, err)
	})
}
