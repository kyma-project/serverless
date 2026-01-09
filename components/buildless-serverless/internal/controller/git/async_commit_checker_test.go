package git

import (
	"context"
	"testing"
	"time"

	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

func TestAsyncLastCommitChecker(t *testing.T) {
	t.Run("order last commit check and cleanup cache entry", func(t *testing.T) {
		repo := "test-repo"
		ref := "test-ref"
		auth := &GitAuth{
			secretName:      "test",
			secretNamespace: "default",
		}
		checker := asyncLastCommitChecker{
			log:                zap.NewNop().Sugar(),
			cacheEntryLifespan: time.Microsecond,
			getLastCommit: func(repo, ref string, auth *GitAuth) (string, error) {
				return "test-commit", nil
			},
		}

		isOrdered := checker.IsLastCommitCheckOrdered(repo, ref, auth)
		require.False(t, isOrdered, "commit check should not be ordered yet")

		checker.OrderLastCommitCheck(context.Background(), repo, ref, auth)
		isOrdered = checker.IsLastCommitCheckOrdered(repo, ref, auth)
		require.True(t, isOrdered, "commit check should be ordered")

		// wait for cache entry to expire
		time.Sleep(2 * time.Millisecond)

		isOrdered = checker.IsLastCommitCheckOrdered(repo, ref, auth)
		require.False(t, isOrdered, "commit check order should be cleaned up after lifespan")
	})

	t.Run("remove order manually after getting error result", func(t *testing.T) {
		repo := "test-repo"
		ref := "test-ref"
		auth := &GitAuth{
			secretName: "test",
		}

		checker := asyncLastCommitChecker{
			log:                zap.NewNop().Sugar(),
			cacheEntryLifespan: time.Minute,
			getLastCommit: func(repo, ref string, auth *GitAuth) (string, error) {
				return "", errors.New("test error")
			},
		}

		checker.OrderLastCommitCheck(context.Background(), repo, ref, auth)

		// wait for cache entry to expire
		time.Sleep(2 * time.Millisecond)

		result := checker.GetLastCommitCheckResult(repo, ref, auth)
		require.NotNil(t, result)
		require.Empty(t, result.Commit)
		require.ErrorContains(t, result.Error, "test error")

		checker.DeleteLastCommitCheckOrder(repo, ref, auth)

		isOrdered := checker.IsLastCommitCheckOrdered(repo, ref, auth)
		require.False(t, isOrdered, "commit check order should be removed manually")
	})
}
