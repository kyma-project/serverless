package state

import (
	"context"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	apiruntime "k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"testing"

	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	"github.com/kyma-project/serverless-manager/internal/chart"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/types"
)

func Test_buildSFnApplyResources(t *testing.T) {
	t.Run("switch state and add condition when condition is missing", func(t *testing.T) {
		scheme := apiruntime.NewScheme()
		require.NoError(t, v1alpha1.AddToScheme(scheme))
		eventRecorder := record.NewFakeRecorder(5)
		serverless := v1alpha1.Serverless{
			ObjectMeta: v1.ObjectMeta{
				Name:            "test",
				Namespace:       "test-namespace",
				ResourceVersion: "777",
			},
			Status: v1alpha1.ServerlessStatus{},
		}
		s := &systemState{
			instance: serverless,
		}
		r := &reconciler{
			cfg: cfg{
				finalizer: v1alpha1.Finalizer,
			},
			k8s: k8s{
				client: fake.
					NewClientBuilder().
					WithScheme(scheme).
					WithObjects(serverless.DeepCopy()).
					Build(),
				EventRecorder: eventRecorder,
			},
		}

		// return sFnUpdateProcessingState when condition is missing
		stateFn := sFnApplyResources()
		next, result, err := stateFn(nil, r, s)

		//TODO: I don't know how to check next function. Nethods below don't work.
		//expected := sFnRequeue()
		//requireEqualFunc(t, expected, next)
		require.NotNil(t, next)
		require.Nil(t, result)
		require.Nil(t, err)

		require.Equal(t, v1alpha1.StateProcessing, s.instance.Status.State)
		require.Len(t, s.instance.Status.Conditions, 1)
		condition := s.instance.Status.Conditions[0]
		require.Equal(t, string(v1alpha1.ConditionTypeInstalled), condition.Type)
		require.Equal(t, string(v1alpha1.ConditionReasonInstallation), condition.Reason)
		require.Equal(t, "Installing for configuration", condition.Message)
	})

	t.Run("apply resources", func(t *testing.T) {
		s := &systemState{
			instance: testInstalledServerless,
			chartConfig: &chart.Config{
				Cache: fixEmptyManifestCache(),
				CacheKey: types.NamespacedName{
					Name:      testInstalledServerless.GetName(),
					Namespace: testInstalledServerless.GetNamespace(),
				},
				Release: chart.Release{
					Name:      testInstalledServerless.GetName(),
					Namespace: testInstalledServerless.GetNamespace(),
				},
			},
		}

		r := &reconciler{}

		// return sFnApplyResources
		stateFn := sFnApplyResources()
		requireEqualFunc(t, sFnApplyResources(), stateFn)

		// run installation process and return verificating state
		next, result, err := stateFn(context.Background(), r, s)

		expectedNext := sFnVerifyResources()

		requireEqualFunc(t, expectedNext, next)
		require.Nil(t, result)
		require.Nil(t, err)
	})

	t.Run("install chart error", func(t *testing.T) {
		s := &systemState{
			instance: testInstalledServerless,
			chartConfig: &chart.Config{
				Cache: fixManifestCache("\t"),
				CacheKey: types.NamespacedName{
					Name:      testInstalledServerless.GetName(),
					Namespace: testInstalledServerless.GetNamespace(),
				},
			},
		}

		r := &reconciler{
			log: zap.NewNop().Sugar(),
		}

		// return sFnApplyResources
		stateFn := sFnApplyResources()
		requireEqualFunc(t, sFnApplyResources(), stateFn)

		// handle error and return update condition state
		next, result, err := stateFn(context.Background(), r, s)

		expectedNext := sFnUpdateErrorState(
			v1alpha1.ConditionTypeInstalled,
			v1alpha1.ConditionReasonInstallationErr,
			err,
		)

		requireEqualFunc(t, expectedNext, next)
		require.Nil(t, result)
		require.Nil(t, err)
	})
}
