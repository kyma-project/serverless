package state

import (
	"context"
	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"
)

func Test_sFnAddFinalizer(t *testing.T) {
	t.Run("set finalizer", func(t *testing.T) {
		scheme := runtime.NewScheme()
		require.NoError(t, v1alpha1.AddToScheme(scheme))

		serverless := v1alpha1.Serverless{
			ObjectMeta: v1.ObjectMeta{
				Name:            "test-name",
				Namespace:       "test-namespace",
				ResourceVersion: "123",
			},
		}
		s := &systemState{
			instance: serverless,
		}

		r := &reconciler{
			cfg: cfg{
				finalizer: v1alpha1.Finalizer,
			},
			k8s: k8s{
				client: fake.NewClientBuilder().
					WithScheme(scheme).
					WithObjects(&serverless).
					Build(),
			},
		}

		// set finalizer
		next, result, err := sFnAddFinalizer(context.Background(), r, s)
		require.Nil(t, result)
		require.NoError(t, err)
		requireEqualFunc(t, sFnInitialize, next)

		// check finalizer
		require.Contains(t, s.instance.GetFinalizers(), r.cfg.finalizer)

		//TODO: test kubernetes object
	})

	t.Run("stop when no finalizer and instance is being deleted", func(t *testing.T) {
		r := &reconciler{
			cfg: cfg{
				finalizer: v1alpha1.Finalizer,
			},
		}

		metaTimeNow := v1.Now()
		s := &systemState{
			instance: v1alpha1.Serverless{
				ObjectMeta: v1.ObjectMeta{
					DeletionTimestamp: &metaTimeNow,
				},
			},
		}

		// stop
		next, result, err := sFnAddFinalizer(context.Background(), r, s)
		require.Nil(t, next)
		require.Nil(t, result)
		require.Nil(t, err)
	})
}
