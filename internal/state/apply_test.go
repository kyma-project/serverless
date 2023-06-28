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
	t.Run("switch state and add condition when condition is missing", func(t *testing.T) {
		s := &systemState{
			instance: v1alpha1.Serverless{},
		}

		// return sFnUpdateProcessingState when condition is missing
		next, result, err := sFnApplyResources(nil, nil, s)

		expected := sFnUpdateStatusAndRequeue
		requireEqualFunc(t, expected, next)
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

		// run installation process and return verificating state
		next, result, err := sFnApplyResources(context.Background(), r, s)

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

		// handle error and return update condition state
		next, result, err := sFnApplyResources(context.Background(), r, s)

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
