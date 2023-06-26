package state

import (
	"context"
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
	t.Run("switch state when condition is missing", func(t *testing.T) {
		eventRecorder := record.NewFakeRecorder(5)
		s := &systemState{
			instance: v1alpha1.Serverless{},
		}
		r := &reconciler{
			cfg: cfg{
				finalizer: v1alpha1.Finalizer,
			},
			k8s: k8s{
				client:        fake.NewClientBuilder().Build(),
				EventRecorder: eventRecorder,
			},
		}

		// return sFnUpdateProcessingState when condition is missing
		stateFn := sFnApplyResources()
		next, result, err := stateFn(nil, r, s)

		expected := sFnRequeue()

		requireEqualFunc(t, expected, next)
		require.Nil(t, result)
		require.Nil(t, err)
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
