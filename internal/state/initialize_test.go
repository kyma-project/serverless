package state

import (
	"context"
	"testing"

	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func Test_sFnInitialize(t *testing.T) {
	t.Run("set finalizer", func(t *testing.T) {
		s := &systemState{
			instance: v1alpha1.Serverless{},
		}

		r := &reconciler{
			cfg: cfg{
				finalizer: v1alpha1.Finalizer,
			},
			k8s: k8s{
				client: fake.NewClientBuilder().Build(),
			},
		}

		// set finalizer
		next, result, err := sFnInitialize(context.Background(), r, s)
		require.Nil(t, next) // expected because client is not fully setup
		require.Equal(t, &ctrl.Result{Requeue: true}, result)
		require.Error(t, err)

		// check finalizer
		require.Contains(t, s.instance.GetFinalizers(), r.cfg.finalizer)
	})

	t.Run("stop when no finalizer and instance is being deleted", func(t *testing.T) {
		r := &reconciler{
			cfg: cfg{
				finalizer: v1alpha1.Finalizer,
			},
		}

		metaTimeNow := metav1.Now()
		s := &systemState{
			instance: v1alpha1.Serverless{
				ObjectMeta: metav1.ObjectMeta{
					DeletionTimestamp: &metaTimeNow,
				},
			},
		}

		// stop
		next, result, err := sFnInitialize(context.Background(), r, s)
		require.Nil(t, next)
		require.Nil(t, result)
		require.Nil(t, err)
	})

	t.Run("setup and return next step sFnRegistryConfiguration", func(t *testing.T) {
		r := &reconciler{
			cfg: cfg{
				finalizer: v1alpha1.Finalizer,
			},
			k8s: k8s{
				client: fake.NewClientBuilder().Build(),
			},
		}

		s := &systemState{
			instance: v1alpha1.Serverless{
				ObjectMeta: metav1.ObjectMeta{
					Finalizers: []string{
						r.cfg.finalizer,
					},
				},
				Spec: v1alpha1.ServerlessSpec{},
			},
		}

		// setup and return buildSFnPrerequisites
		next, result, err := sFnInitialize(context.Background(), r, s)

		expectedNext := sFnRegistryConfiguration
		requireEqualFunc(t, expectedNext, next)
		require.Nil(t, result)
		require.Nil(t, err)
	})

	t.Run("setup and return next step sFnDeleteResources", func(t *testing.T) {
		r := &reconciler{
			cfg: cfg{
				finalizer: v1alpha1.Finalizer,
			},
			k8s: k8s{
				client: fake.NewClientBuilder().Build(),
			},
		}

		metaTine := metav1.Now()
		s := &systemState{
			instance: v1alpha1.Serverless{
				ObjectMeta: metav1.ObjectMeta{
					Finalizers: []string{
						r.cfg.finalizer,
					},
					DeletionTimestamp: &metaTine,
				},
				Spec: v1alpha1.ServerlessSpec{},
			},
		}

		// setup and return buildSFnDeleteResources
		next, result, err := sFnInitialize(context.Background(), r, s)

		expectedNext := sFnDeleteResources
		requireEqualFunc(t, expectedNext, next)
		require.Nil(t, result)
		require.Nil(t, err)
	})

}
