package state

import (
	"context"
	"github.com/kyma-project/serverless/components/operator/api/v1alpha1"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
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
		require.NoError(t, err)
		require.Nil(t, result)
		requireEqualFunc(t, sFnInitialize, next)

		// check finalizer in systemState
		require.Contains(t, s.instance.GetFinalizers(), r.cfg.finalizer)

		// check finalizer in k8s
		obj := v1alpha1.Serverless{}
		err = r.k8s.client.Get(context.Background(),
			client.ObjectKey{
				Namespace: serverless.Namespace,
				Name:      serverless.Name,
			},
			&obj)
		require.NoError(t, err)
		require.Contains(t, obj.GetFinalizers(), r.cfg.finalizer)
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
		require.Nil(t, err)
		require.Nil(t, result)
		require.Nil(t, next)
	})
}
