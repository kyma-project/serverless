package state

import (
	"testing"

	"github.com/kyma-project/serverless-manager/internal/chart"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func Test_sFnDeleteResources(t *testing.T) {
	t.Run("cascade deletion", func(t *testing.T) {
		stateFn := deletionStrategyBuilder(cascadeDeletionStrategy)

		s := &systemState{
			instance: testInstalledServerless,
			chartConfig: &chart.Config{
				Cache: testEmptyManifestCache(),
				Release: chart.Release{
					Name:      testInstalledServerless.GetName(),
					Namespace: testInstalledServerless.GetNamespace(),
				},
			},
		}

		r := &reconciler{}

		next, result, err := stateFn(nil, r, s)

		expectedNext := sFnUpdateDeletingState(
			sFnRemoveFinalizer(),
			"Normal",
			"Deleted",
			"Serverless module deleted",
		)

		requireEqualFunc(t, expectedNext, next)
		require.Nil(t, result)
		require.Nil(t, err)
	})

	t.Run("upstream deletion error", func(t *testing.T) {
		stateFn := deletionStrategyBuilder(upstreamDeletionStrategy)

		s := &systemState{
			instance: testInstalledServerless,
			chartConfig: &chart.Config{
				Cache:   testEmptyManifestCache(),
				Release: chart.Release{},
			},
		}

		r := &reconciler{
			log: zap.NewNop().Sugar(),
		}

		next, result, err := stateFn(nil, r, s)

		expectedNext := sFnUpdateDeletingState(
			sFnRequeue(),
			"Warning",
			"Deletion",
			"test error",
		)

		requireEqualFunc(t, expectedNext, next)
		require.Nil(t, result)
		require.Nil(t, err)
	})

	t.Run("safe deletion error while checking orphan resources", func(t *testing.T) {
		wrongStrategy := deletionStrategy("test-strategy")
		stateFn := deletionStrategyBuilder(wrongStrategy)

		s := &systemState{
			instance: testInstalledServerless,
			chartConfig: &chart.Config{
				Cache:   testEmptyManifestCache(),
				Release: chart.Release{},
			},
		}

		r := &reconciler{
			log: zap.NewNop().Sugar(),
		}

		next, result, err := stateFn(nil, r, s)

		expectedNext := sFnUpdateDeletingState(
			sFnRequeue(),
			"Warning",
			"Deletion",
			"test error",
		)

		requireEqualFunc(t, expectedNext, next)
		require.Nil(t, result)
		require.Nil(t, err)
	})

	t.Run("safe deletion", func(t *testing.T) {
		wrongStrategy := deletionStrategy("test-strategy")
		stateFn := deletionStrategyBuilder(wrongStrategy)

		s := &systemState{
			instance: testInstalledServerless,
			chartConfig: &chart.Config{
				Cache: testEmptyManifestCache(),
				Release: chart.Release{
					Name:      testInstalledServerless.GetName(),
					Namespace: testInstalledServerless.GetNamespace(),
				},
			},
		}

		r := &reconciler{
			log: zap.NewNop().Sugar(),
		}

		next, result, err := stateFn(nil, r, s)

		expectedNext := sFnUpdateDeletingState(
			sFnRemoveFinalizer(),
			"Normal",
			"Deleted",
			"Serverless module deleted",
		)

		requireEqualFunc(t, expectedNext, next)
		require.Nil(t, result)
		require.Nil(t, err)
	})
}
