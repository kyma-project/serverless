package state

import (
	"testing"

	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/pointer"
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
				client: fake.NewFakeClient(),
			},
		}

		// set finalizer
		next, result, err := sFnInitialize(nil, r, s)
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
		next, result, err := sFnInitialize(nil, r, s)
		require.Nil(t, next)
		require.Nil(t, result)
		require.Nil(t, err)
	})

	t.Run("setup error", func(t *testing.T) {
		r := &reconciler{
			log: zap.NewNop().Sugar(),
			cfg: cfg{
				finalizer: v1alpha1.Finalizer,
			},
			k8s: k8s{
				client: fake.NewFakeClient(),
			},
		}

		s := &systemState{
			instance: v1alpha1.Serverless{
				ObjectMeta: metav1.ObjectMeta{
					Finalizers: []string{
						r.cfg.finalizer,
					},
				},
				Spec: v1alpha1.ServerlessSpec{
					DockerRegistry: &v1alpha1.DockerRegistry{
						SecretName: pointer.String("does-not-exist"),
					},
				},
			},
		}

		// stop
		next, result, err := sFnInitialize(nil, r, s)

		expectedNext := sFnUpdateStatusAndStop
		requireEqualFunc(t, expectedNext, next)
		require.Nil(t, result)
		require.Nil(t, err)

		status := s.instance.Status
		require.Equal(t, v1alpha1.StateError, status.State)
		requireContainsCondition(t, status,
			v1alpha1.ConditionTypeConfigured,
			v1alpha1.ConditionReasonConfigurationErr,
			"resolving manifest failed: secrets \"does-not-exist\" not found",
		)
	})

	t.Run("setup and return next step sFnOptionalDependencies", func(t *testing.T) {
		r := &reconciler{
			cfg: cfg{
				finalizer: v1alpha1.Finalizer,
			},
			k8s: k8s{
				client: fake.NewFakeClient(),
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
		next, result, err := sFnInitialize(nil, r, s)

		expectedNext := sFnOptionalDependencies
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
				client: fake.NewFakeClient(),
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
		next, result, err := sFnInitialize(nil, r, s)

		expectedNext := sFnDeleteResources
		requireEqualFunc(t, expectedNext, next)
		require.Nil(t, result)
		require.Nil(t, err)
	})

	t.Run("take snapshot", func(t *testing.T) {
		r := &reconciler{
			cfg: cfg{
				finalizer: v1alpha1.Finalizer,
			},
			k8s: k8s{
				client: fake.NewFakeClient(),
			},
		}
		serverless := v1alpha1.Serverless{
			ObjectMeta: metav1.ObjectMeta{
				Finalizers: []string{
					r.cfg.finalizer,
				},
			},
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

		_, _, err := sFnInitialize(nil, r, s)
		require.NoError(t, err)

		// check status
		require.Equal(t, serverless.Status, s.snapshot)
	})
}
