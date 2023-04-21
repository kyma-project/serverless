package state

import (
	"errors"
	"testing"

	"github.com/kyma-project/serverless-manager/api/v1alpha1"
	"github.com/kyma-project/serverless-manager/internal/chart"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

var (
	testDeletingServerless = func() v1alpha1.Serverless {
		serverless := testInstalledServerless
		serverless.Status.State = v1alpha1.StateDeleting
		return serverless
	}()
)

func Test_sFnDeleteResources(t *testing.T) {
	t.Run("update condition", func(t *testing.T) {
		s := &systemState{
			instance: v1alpha1.Serverless{},
		}

		stateFn := sFnDeleteResources()
		next, result, err := stateFn(nil, nil, s)

		expectedNext := sFnUpdateDeletingState(
			"Deletion",
			"Uninstalling",
		)

		requireEqualFunc(t, expectedNext, next)
		require.Nil(t, result)
		require.Nil(t, err)
	})
	t.Run("choose deletion strategy", func(t *testing.T) {
		s := &systemState{
			instance: testDeletingServerless,
		}

		stateFn := sFnDeleteResources()
		next, result, err := stateFn(nil, nil, s)

		expectedNext := deletionStrategyBuilder(defaultDeletionStrategy)

		requireEqualFunc(t, expectedNext, next)
		require.Nil(t, result)
		require.Nil(t, err)
	})
	t.Run("cascade deletion", func(t *testing.T) {
		stateFn := deletionStrategyBuilder(cascadeDeletionStrategy)

		s := &systemState{
			instance: testDeletingServerless,
			chartConfig: &chart.Config{
				Cache: testEmptyManifestCache(),
				Release: chart.Release{
					Name:      testDeletingServerless.GetName(),
					Namespace: testDeletingServerless.GetNamespace(),
				},
			},
		}

		r := &reconciler{}

		next, result, err := stateFn(nil, r, s)

		expectedNext := sFnRemoveFinalizer()

		requireEqualFunc(t, expectedNext, next)
		require.Nil(t, result)
		require.Nil(t, err)
	})

	t.Run("upstream deletion error", func(t *testing.T) {
		stateFn := deletionStrategyBuilder(upstreamDeletionStrategy)

		s := &systemState{
			instance: testDeletingServerless,
			chartConfig: &chart.Config{
				Cache:   testEmptyManifestCache(),
				Release: chart.Release{},
			},
		}

		r := &reconciler{
			log: zap.NewNop().Sugar(),
		}

		next, result, err := stateFn(nil, r, s)

		expectedNext := sFnUpdateDeletingErrorState(
			"Deletion",
			errors.New("test error"),
		)

		requireEqualFunc(t, expectedNext, next)
		require.Nil(t, result)
		require.Nil(t, err)
	})

	t.Run("safe deletion error while checking orphan resources", func(t *testing.T) {
		wrongStrategy := deletionStrategy("test-strategy")
		stateFn := deletionStrategyBuilder(wrongStrategy)

		s := &systemState{
			instance: testDeletingServerless,
			chartConfig: &chart.Config{
				Cache:   testEmptyManifestCache(),
				Release: chart.Release{},
			},
		}

		r := &reconciler{
			log: zap.NewNop().Sugar(),
		}

		next, result, err := stateFn(nil, r, s)

		expectedNext := sFnUpdateDeletingErrorState(
			"Deletion",
			errors.New("test error"),
		)

		requireEqualFunc(t, expectedNext, next)
		require.Nil(t, result)
		require.Nil(t, err)
	})

	t.Run("safe deletion", func(t *testing.T) {
		wrongStrategy := deletionStrategy("test-strategy")
		stateFn := deletionStrategyBuilder(wrongStrategy)

		s := &systemState{
			instance: testDeletingServerless,
			chartConfig: &chart.Config{
				Cache: testEmptyManifestCache(),
				Release: chart.Release{
					Name:      testDeletingServerless.GetName(),
					Namespace: testDeletingServerless.GetNamespace(),
				},
			},
		}

		r := &reconciler{
			log: zap.NewNop().Sugar(),
		}

		next, result, err := stateFn(nil, r, s)

		expectedNext := sFnRemoveFinalizer()

		requireEqualFunc(t, expectedNext, next)
		require.Nil(t, result)
		require.Nil(t, err)
	})
}
