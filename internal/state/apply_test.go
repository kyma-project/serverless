package state

import (
	"context"
	"testing"

	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	"github.com/kyma-project/serverless-manager/internal/chart"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"k8s.io/apimachinery/pkg/types"
)

func Test_buildSFnApplyResources(t *testing.T) {
	t.Run("switch state when condition is missing", func(t *testing.T) {
		s := &systemState{
			instance: v1alpha1.Serverless{},
		}

		// return sFnUpdateProcessingState when condition is missing
		stateFn := sFnApplyResources()
		next, result, err := stateFn(nil, nil, s)

		expected := sFnUpdateProcessingState(
			v1alpha1.ConditionTypeInstalled,
			v1alpha1.ConditionReasonInstallation,
			"Installing for configuration",
		)

		requireEqualFunc(t, expected, next)
		require.Nil(t, result)
		require.Nil(t, err)
	})

	t.Run("apply resources", func(t *testing.T) {
		s := &systemState{
			instance: testInstalledServerless,
			chartConfig: &chart.Config{
				Cache: testEmptyManifestCache(),
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
				Cache: testEmptyManifestCache(),
				CacheKey: types.NamespacedName{
					Name:      "does-not-exist",
					Namespace: "test",
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
