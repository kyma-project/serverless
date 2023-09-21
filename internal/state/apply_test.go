package state

import (
	"context"
	"testing"

	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	"github.com/kyma-project/serverless-manager/internal/chart"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

func Test_buildSFnApplyResources(t *testing.T) {
	t.Run("switch state and add condition when condition is missing", func(t *testing.T) {
		s := &systemState{
			instance: v1alpha1.Serverless{},
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

		next, result, err := sFnApplyResources(context.Background(), nil, s)

		expected := sFnVerifyResources
		requireEqualFunc(t, expected, next)
		require.Nil(t, result)
		require.Nil(t, err)

		status := s.instance.Status
		require.Equal(t, v1alpha1.StateProcessing, status.State)
		requireContainsCondition(t, status,
			v1alpha1.ConditionTypeInstalled,
			metav1.ConditionUnknown,
			v1alpha1.ConditionReasonInstallation,
			"Installing for configuration",
		)
	})

	t.Run("apply resources", func(t *testing.T) {
		s := &systemState{
			instance: *testInstalledServerless.DeepCopy(),
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

		expectedNext := sFnVerifyResources
		requireEqualFunc(t, expectedNext, next)
		require.Nil(t, result)
		require.Nil(t, err)
	})

	t.Run("install chart error", func(t *testing.T) {
		s := &systemState{
			instance: *testInstalledServerless.DeepCopy(),
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

		require.Nil(t, next)
		require.Nil(t, result)
		require.EqualError(t, err, "could not parse chart manifest: yaml: found character that cannot start any token")

		status := s.instance.Status
		require.Equal(t, v1alpha1.StateError, status.State)
		requireContainsCondition(t, status,
			v1alpha1.ConditionTypeInstalled,
			metav1.ConditionFalse,
			v1alpha1.ConditionReasonInstallationErr,
			"could not parse chart manifest: yaml: found character that cannot start any token",
		)
	})
}
