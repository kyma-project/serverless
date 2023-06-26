package state

import (
	"errors"
	"testing"

	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	"github.com/kyma-project/serverless-manager/internal/chart"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var (
	testDeletingServerless = func() v1alpha1.Serverless {
		serverless := testInstalledServerless
		serverless.Status.State = v1alpha1.StateDeleting
		serverless.Status.Conditions = []metav1.Condition{
			{
				Type:   string(v1alpha1.ConditionTypeDeleted),
				Reason: string(v1alpha1.ConditionReasonDeletion),
				Status: metav1.ConditionUnknown,
			},
		}
		return serverless
	}()
)

func Test_sFnDeleteResources(t *testing.T) {
	t.Run("update condition", func(t *testing.T) {
		s := &systemState{
			instance: v1alpha1.Serverless{},
		}

		next, result, err := sFnDeleteResources(nil, nil, s)

		expectedNext := sFnUpdateStatusAndRequeue
		requireEqualFunc(t, expectedNext, next)
		require.Nil(t, result)
		require.Nil(t, err)

		status := s.instance.Status
		require.Equal(t, v1alpha1.StateDeleting, status.State)
		requireContainsCondition(t, status,
			v1alpha1.ConditionTypeDeleted,
			v1alpha1.ConditionReasonDeletion,
			"Uninstalling",
		)
	})
	t.Run("choose deletion strategy", func(t *testing.T) {
		s := &systemState{
			instance: *testDeletingServerless.DeepCopy(),
		}

		next, result, err := sFnDeleteResources(nil, nil, s)

		expectedNext := deletionStrategyBuilder(defaultDeletionStrategy)
		requireEqualFunc(t, expectedNext, next)
		require.Nil(t, result)
		require.Nil(t, err)
	})
	t.Run("cascade deletion", func(t *testing.T) {
		stateFn := deletionStrategyBuilder(cascadeDeletionStrategy)

		s := &systemState{
			instance: *testDeletingServerless.DeepCopy(),
			chartConfig: &chart.Config{
				Cache: fixEmptyManifestCache(),
				CacheKey: types.NamespacedName{
					Name:      testDeletingServerless.GetName(),
					Namespace: testDeletingServerless.GetNamespace(),
				},
			},
		}

		r := &reconciler{}

		next, result, err := stateFn(nil, r, s)

		expectedNext := sFnUpdateStatusAndRequeue
		requireEqualFunc(t, expectedNext, next)
		require.Nil(t, result)
		require.Nil(t, err)

		status := s.instance.Status
		require.Equal(t, v1alpha1.StateDeleting, status.State)
		requireContainsCondition(t, status,
			v1alpha1.ConditionTypeDeleted,
			v1alpha1.ConditionReasonDeleted,
			"Serverless module deleted",
		)
	})

	t.Run("upstream deletion error", func(t *testing.T) {
		stateFn := deletionStrategyBuilder(upstreamDeletionStrategy)

		s := &systemState{
			instance: *testDeletingServerless.DeepCopy(),
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

		next, result, err := stateFn(nil, r, s)

		expectedNext := sFnUpdateStatusWithError(errors.New("anything"))
		requireEqualFunc(t, expectedNext, next)
		require.Nil(t, result)
		require.Nil(t, err)

		status := s.instance.Status
		require.Equal(t, v1alpha1.StateError, status.State)
		requireContainsCondition(t, status,
			v1alpha1.ConditionTypeDeleted,
			v1alpha1.ConditionReasonDeletionErr,
			"could not parse chart manifest: yaml: found character that cannot start any token",
		)
	})

	t.Run("safe deletion error while checking orphan resources", func(t *testing.T) {
		wrongStrategy := deletionStrategy("test-strategy")
		stateFn := deletionStrategyBuilder(wrongStrategy)

		s := &systemState{
			instance: *testDeletingServerless.DeepCopy(),
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

		next, result, err := stateFn(nil, r, s)

		expectedNext := sFnUpdateStatusWithError(errors.New("anything"))
		requireEqualFunc(t, expectedNext, next)
		require.Nil(t, result)
		require.Nil(t, err)

		status := s.instance.Status
		require.Equal(t, v1alpha1.StateError, status.State)
		requireContainsCondition(t, status,
			v1alpha1.ConditionTypeDeleted,
			v1alpha1.ConditionReasonDeletionErr,
			"could not parse chart manifest: yaml: found character that cannot start any token",
		)
	})

	t.Run("safe deletion", func(t *testing.T) {
		wrongStrategy := deletionStrategy("test-strategy")
		stateFn := deletionStrategyBuilder(wrongStrategy)

		s := &systemState{
			instance: *testDeletingServerless.DeepCopy(),
			chartConfig: &chart.Config{
				Cache: fixEmptyManifestCache(),
				CacheKey: types.NamespacedName{
					Name:      testDeletingServerless.GetName(),
					Namespace: testDeletingServerless.GetNamespace(),
				},
			},
		}

		r := &reconciler{
			log: zap.NewNop().Sugar(),
		}

		next, result, err := stateFn(nil, r, s)

		expectedNext := sFnUpdateStatusAndRequeue
		requireEqualFunc(t, expectedNext, next)
		require.Nil(t, result)
		require.Nil(t, err)

		status := s.instance.Status
		require.Equal(t, v1alpha1.StateDeleting, status.State)
		requireContainsCondition(t, status,
			v1alpha1.ConditionTypeDeleted,
			v1alpha1.ConditionReasonDeleted,
			"Serverless module deleted",
		)
	})
}
